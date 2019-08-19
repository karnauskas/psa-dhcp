package client

import (
	"bytes"
	"context"
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/client/arpping"
	"gitlab.com/adrian_blx/psa-dhcp/lib/libif"
)

// runStateInitIface removes any IPv4 configuration from the interface and brings it up.
func (dx *dclient) runStatePurgeInterface(nextState int) {
	dx.l.Printf("unconfiguring interface\n")
	if err := libif.Unconfigure(dx.iface); err != nil {
		dx.l.Printf("Unconfigure returned error %v\n", err)
	}
	if err := libif.Up(dx.iface); err != nil {
		dx.l.Printf("Bringing up interface returned error %v\n", err)
	}
	dx.state = nextState
	dx.callback(nil)
}

// runStateIfconfig applies the current state of the client to the network interface.
func (dx *dclient) runStateIfconfig(nextState int) {
	nc := dx.buildNetconfig()
	dx.l.Printf("Configuring interface to use IP %s/%s via %s\n", nc.IP, nc.Netmask, nc.Router)
	if err := libif.SetIface(nc); err != nil {
		dx.panicReset("Unexpected error while configuring interface, falling back to INIT in 30 sec! (error was: %v)\n", err)
	} else {
		dx.state = nextState
		dx.callback(&nc)
	}
}

// runStateArpCheck performs an arp ping on our IP to validate it is unused.
func (dx *dclient) runStateArpCheck(nextState int) {
	dx.l.Printf("Running ARPING for %s\n", dx.lastMsg.YourIP)
	mac, err := arpping.Ping(dx.ctx, dx.iface, net.IPv4(0, 0, 0, 0), dx.lastMsg.YourIP)
	if err == nil && !bytes.Equal(mac, dx.iface.HardwareAddr) {
		dx.panicReset("IP %v is already in use by %v", dx.lastMsg.YourIP, mac)
	} else {
		dx.state = nextState
	}
}

// panicReset unconfigures the service after some time.
func (dx *dclient) panicReset(f string, args ...interface{}) {
	dx.l.Printf("PANIC RESET: "+f, args...)

	libif.Unconfigure(dx.iface) // drop our own IP; best effort.
	fctx, _ := context.WithTimeout(dx.ctx, time.Second*30)
	<-fctx.Done()
	dx.state = statePurgeInterface
}
