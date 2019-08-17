package client

import (
	"bytes"
	"context"
	"log"
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/client/arpping"
	"gitlab.com/adrian_blx/psa-dhcp/lib/client/msgtmpl"
	vy "gitlab.com/adrian_blx/psa-dhcp/lib/client/verify"
	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
	"gitlab.com/adrian_blx/psa-dhcp/lib/libif"
)

const (
	// Unconfigures the interface and brings it up.
	stateInitIface = iota
	// Send initial DHCPDISCOVER.
	stateInit
	// Selects a dhcp server via DHCPREQUEST.
	stateSelecting
	// Verify the current config by sending an ARP ping
	stateArpCheck
	// Configures the OS with the received configuration.
	stateIfconfig
	// We have a lease and are sleeping.
	stateBound
	// We try to renew the state (unicast)
	stateRenewing
	// We try to rebind (broadcast)
	stateRebinding
)

type boundDeadlines struct {
	// t1 is the time at which the client enters stateRenewing.
	t1 time.Time
	// t2 is the time at which the client enters stateRebinding.
	t2 time.Time
	// tx is the time at which we give up our IP.
	tx time.Time
}

type dclient struct {
	ctx            context.Context
	l              *log.Logger
	iface          *net.Interface
	state          int
	lastMsg        dhcpmsg.Message
	lastOpts       dhcpmsg.DecodedOptions
	boundDeadlines boundDeadlines
}

type vrfyFunc func(dhcpmsg.Message, dhcpmsg.DecodedOptions) bool

func New(ctx context.Context, l *log.Logger, iface *net.Interface) *dclient {
	return &dclient{ctx: ctx, l: l, iface: iface, state: stateInitIface}
}

func (dx *dclient) Run() error {
	for {
		switch dx.state {
		case stateInitIface:
			dx.runStateInitIface()
		case stateInit:
			dx.runStateInit()
		case stateSelecting:
			dx.runStateSelecting()
		case stateArpCheck:
			dx.runStateArpCheck()
		case stateIfconfig:
			dx.runStateIfconfig()
		case stateBound:
			now := time.Now()
			dx.boundDeadlines = boundDeadlines{
				t1: now.Add(time.Duration(float64(dx.lastOpts.IPAddressLeaseTime) * 0.5)),
				t2: now.Add(time.Duration(float64(dx.lastOpts.IPAddressLeaseTime) * 0.875)),
				tx: now.Add(dx.lastOpts.IPAddressLeaseTime),
			}
			if dx.lastOpts.RenewalTime > time.Minute &&
				dx.lastOpts.RebindTime > dx.lastOpts.RenewalTime &&
				dx.lastOpts.RebindTime < dx.lastOpts.IPAddressLeaseTime {
				dx.boundDeadlines.t1 = now.Add(dx.lastOpts.RenewalTime)
				dx.boundDeadlines.t2 = now.Add(dx.lastOpts.RebindTime)
			}
			dx.l.Printf("-> Reached BOUND state. Will sleep until T1 expires.")
			dx.l.Printf("T1 = %s", dx.boundDeadlines.t1)
			dx.l.Printf("T2 = %s", dx.boundDeadlines.t2)
			dx.l.Printf("TX = %s", dx.boundDeadlines.tx)
			hackAbsoluteSleep(dx.ctx, dx.boundDeadlines.t1)
			dx.state = stateRenewing
		case stateRenewing:
			dx.l.Printf("-> Reached RENEWING state. Will try until %s", dx.boundDeadlines.t2)
			rq, xid := msgtmpl.RequestRenewing(dx.iface, dx.lastMsg.YourIP, dx.lastOpts.ServerIdentifier)
			if lm, lo, p := dx.advanceState(dx.boundDeadlines.t2, vy.VerifyRenewingAck(dx.lastMsg, dx.lastOpts, xid), rq); p {
				dx.state = stateArpCheck
				dx.lastMsg = lm
				dx.lastOpts = lo
			} else {
				dx.state = stateRebinding
			}
		case stateRebinding:
			dx.l.Printf("-> Reached REBINDING state. Will try until %s", dx.boundDeadlines.tx)
			rq, xid := msgtmpl.RequestRebinding(dx.iface, dx.lastMsg.YourIP)
			if lm, lo, p := dx.advanceState(dx.boundDeadlines.tx, vy.VerifyRebindingAck(dx.lastMsg, dx.lastOpts, xid), rq); p {
				dx.state = stateArpCheck
				dx.lastMsg = lm
				dx.lastOpts = lo
			} else {
				dx.state = stateInit
			}
		default:
			dx.l.Panicf("invalid state: %d\n", dx.state)
		}

		// break if main context is done.
		if err := dx.ctx.Err(); err != nil {
			return err
		}
	}
}

func (dx dclient) currentNetconfig() libif.Ifconfig {
	netmask := dx.lastMsg.YourIP.DefaultMask()
	if _, bits := dx.lastOpts.SubnetMask.Size(); bits != 0 {
		netmask = dx.lastOpts.SubnetMask
	}

	cidr, _ := netmask.Size()

	c := libif.Ifconfig{
		Interface:     dx.iface,
		Router:        dx.lastOpts.Routers[0],
		IP:            dx.lastMsg.YourIP,
		Cidr:          cidr,
		LeaseDuration: dx.lastOpts.IPAddressLeaseTime,
	}
	return c
}

func hackAbsoluteSleep(ctx context.Context, when time.Time) {
	for {
		if ctx.Err() != nil {
			break
		}
		if time.Now().After(when) {
			break
		}
		time.Sleep(time.Second * 3)
	}
}

func (dx *dclient) advanceState(deadline time.Time, vrfy vrfyFunc, sender senderFunc) (dhcpmsg.Message, dhcpmsg.DecodedOptions, bool) {
	ctx, cancel := context.WithDeadline(dx.ctx, deadline)
	defer cancel()

	dx.l.Printf("  ==> waiting for valid reply until %s", deadline)
	go sendMessage(ctx, dx.iface, sender)
	msg, opts, err := catchReply(ctx, dx.iface, vrfy)

	if err != nil {
		// If there was an error, wait until the context expires (if we might have a
		// sock setup error) to avoid flooding the line.
		<-ctx.Done()
		return msg, opts, false
	}
	return msg, opts, true
}

// runStateInitIface removes any IPv4 configuration from the interface and brings it up.
func (dx *dclient) runStateInitIface() {
	dx.l.Printf("unconfiguring interface\n")
	if err := libif.Unconfigure(dx.iface); err != nil {
		dx.l.Printf("Unconfigure returned error %v\n", err)
	}
	if err := libif.Up(dx.iface); err != nil {
		dx.l.Printf("Bringing up interface returned error %v\n", err)
	}
	dx.state = stateInit
}

// runStateInit broadcasts a DHCPDISCOVER message on the interface.
func (dx *dclient) runStateInit() {
	dx.l.Printf("Sending DHCPDISCOVER broadcast\n")

	rq, xid := msgtmpl.Discover(dx.iface)
	if lm, lo, p := dx.advanceState(time.Now().Add(10*time.Minute), vy.VerifyOffer(xid), rq); p {
		dx.state = stateSelecting
		dx.lastMsg = lm
		dx.lastOpts = lo
	}
	// else: can't advance to any other state.
}

// runStateSelecting selects a dhcp server by *broadcasting* a DHCPREQUEST.
func (dx *dclient) runStateSelecting() {
	dx.l.Printf("Sending DHCPREQUEST for %s to %s\n", dx.lastMsg.YourIP, dx.lastOpts.ServerIdentifier)

	rq, xid := msgtmpl.RequestSelecting(dx.iface, dx.lastMsg.YourIP, dx.lastOpts.ServerIdentifier)
	if lm, lo, p := dx.advanceState(time.Now().Add(time.Minute), vy.VerifySelectingAck(dx.lastMsg, dx.lastOpts, xid), rq); p {
		dx.state = stateArpCheck
		dx.lastMsg = lm
		dx.lastOpts = lo
	} else {
		dx.state = stateInit
	}
}

// runStateIfconfig applies the current state of the client to the network interface.
func (dx *dclient) runStateIfconfig() {
	nc := dx.currentNetconfig()
	dx.l.Printf("Configuring interface to use IP %s/%d -> %s\n", nc.IP, nc.Cidr, nc.Router)
	if err := libif.SetIface(nc); err != nil {
		dx.panicReset("Unexpected error while configuring interface, falling back to INIT in 30 sec! (error was: %v)\n", err)
	} else {
		dx.state = stateBound
	}
}

func (dx *dclient) runStateArpCheck() {
	dx.l.Printf("Running ARPING for %s\n", dx.lastMsg.YourIP)
	mac, err := arpping.Ping(dx.ctx, dx.iface, net.IPv4(0, 0, 0, 0), dx.lastMsg.YourIP)
	if err == nil && !bytes.Equal(mac, dx.iface.HardwareAddr) {
		dx.panicReset("IP %v is already in use by %v", dx.lastMsg.YourIP, mac)
	} else {
		dx.state = stateIfconfig
	}
}

func (dx *dclient) panicReset(f string, args ...interface{}) {
	dx.l.Printf("PANIC RESET: "+f, args...)
	dx.state = stateInitIface
	fctx, _ := context.WithTimeout(dx.ctx, time.Second*30)
	<-fctx.Done()
}
