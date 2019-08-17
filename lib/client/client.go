package client

import (
	"context"
	"log"
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
	"gitlab.com/adrian_blx/psa-dhcp/lib/libif"
)

const (
	// Invalid value, crashes the server.
	stateInvalid = iota
	// Unconfigures the interface and brings it up.
	statePurgeInterface
	// Send initial DHCPDISCOVER.
	stateDiscovering
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
	return &dclient{ctx: ctx, l: l, iface: iface, state: statePurgeInterface}
}

func (dx *dclient) Run() error {
	for {
		switch dx.state {
		case statePurgeInterface:
			dx.runStatePurgeInterface(stateDiscovering)
		case stateDiscovering:
			dx.runStateDiscovering(stateSelecting)
		case stateSelecting:
			dx.runStateSelecting(stateArpCheck, stateDiscovering)
		case stateArpCheck:
			dx.runStateArpCheck(stateIfconfig)
		case stateIfconfig:
			dx.runStateIfconfig(stateBound)
		case stateBound:
			dx.runStateBound(stateRenewing)
		case stateRenewing:
			dx.runStateRenewing(stateArpCheck, stateRebinding)
		case stateRebinding:
			dx.runStateRebinding(stateArpCheck, statePurgeInterface)
		default:
			dx.l.Panicf("invalid state: %d\n", dx.state)
		}

		// break if main context is done.
		if err := dx.ctx.Err(); err != nil {
			return err
		}
	}
}

func (dx dclient) buildNetconfig() libif.Ifconfig {
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
