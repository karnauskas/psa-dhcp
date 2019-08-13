package client

import (
	"context"
	"log"
	"math/rand"
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/client/msgtmpl"
	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
	"gitlab.com/adrian_blx/psa-dhcp/lib/libif"
)

const (
	stateInitIface = iota
	stateInit
	stateSelecting
	stateIfconfig
	stateBound
	stateRenewing
)

const (
	minLeaseDuration = time.Second * 10
	maxLeaseDuration = time.Hour * 3
)

type dclient struct {
	ctx      context.Context
	l        *log.Logger
	iface    *net.Interface
	state    int
	lastMsg  dhcpmsg.Message
	lastOpts dhcpmsg.DecodedOptions
}

type vrfyFunc func(dhcpmsg.Message, dhcpmsg.DecodedOptions) bool

func New(ctx context.Context, l *log.Logger, iface *net.Interface) *dclient {
	return &dclient{ctx: ctx, l: l, iface: iface, state: stateInitIface}
}

func (dx *dclient) Run() error {
	var pass bool
	for {
		xid := rand.Uint32()
		dx.l.Printf("%s: is now in state %d\n", dx.iface.Name, dx.state)
		switch dx.state {
		case stateInitIface:
			dx.runStateInitIface()
		case stateInit:
			dx.runStateInit()
		case stateSelecting:
			dx.runStateSelecting()
		case stateIfconfig:
			dx.runStateIfconfig()
		case stateBound:
			t1 := time.Now().Add(normalizeLease(dx.lastOpts.RenewalTime))
			dx.l.Printf("%s: Sleeping, T1 is %s\n", dx.iface.Name, t1)
			hackAbsoluteSleep(dx.ctx, t1)
			dx.state = stateRenewing
		case stateRenewing:
			// FIXME: This should be unicast with the correct mac.
			tmpl := msgtmpl.New(dx.iface, xid)
			rq := func() []byte { return tmpl.RequestRenewing(dx.lastMsg.YourIP, dx.lastOpts.ServerIdentifier) }
			dx.lastMsg, dx.lastOpts, pass = dx.advanceState(verifyRenewAck(dx.lastMsg, xid), rq)
			if pass {
				dx.state = stateIfconfig
			} else {
				// should get into other state after T1.
				// not sure how long we will stick around in this one? just loop?
				// Or maybe have RequestRenewing have a deadline? would be best
				// we could then set it to T2.
				dx.state = stateInitIface
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

func normalizeLease(d time.Duration) time.Duration {
	if d < minLeaseDuration {
		return minLeaseDuration
	}
	if d > maxLeaseDuration {
		return maxLeaseDuration
	}
	return d
}

func (dx dclient) currentNetconfig() libif.Ifconfig {
	netmask := dx.lastMsg.YourIP.DefaultMask()
	if _, bits := dx.lastOpts.SubnetMask.Size(); bits != 0 {
		netmask = dx.lastOpts.SubnetMask
	}

	cidr, _ := netmask.Size()

	c := libif.Ifconfig{
		Interface: dx.iface,
		Router:    dx.lastOpts.Routers[0],
		IP:        dx.lastMsg.YourIP,
		Cidr:      cidr,
		// This does not survive sleeps but is still nice to have in case this process dies.
		LeaseDuration: normalizeLease(dx.lastOpts.IPAddressLeaseTime),
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

func (dx *dclient) advanceState(vrfy vrfyFunc, sender senderFunc) (dhcpmsg.Message, dhcpmsg.DecodedOptions, bool) {
	ctx, cancel := context.WithTimeout(dx.ctx, time.Second*60)
	defer cancel()

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

func (dx *dclient) runStateInitIface() {
	dx.l.Printf("%s: unconfiguring interface\n", dx.iface.Name)
	if err := libif.Unconfigure(dx.iface); err != nil {
		dx.l.Printf("%s: unconfigure returned error %v\n", dx.iface.Name, err)
	}
	if err := libif.Up(dx.iface); err != nil {
		dx.l.Printf("%s: bringing up interface returned error %v\n", dx.iface.Name, err)
	}
	dx.state = stateInit
}

func (dx *dclient) runStateInit() {
	dx.l.Printf("%s: Sending DHCPDISCOVER broadcast\n", dx.iface.Name)

	xid := rand.Uint32()
	tmpl := msgtmpl.New(dx.iface, xid)
	var pass bool
	dx.lastMsg, dx.lastOpts, pass = dx.advanceState(verifyOffer(xid), func() []byte { return tmpl.Discover() })
	if pass {
		dx.state = stateSelecting
	}
	// else: can't advance to any other state.
}

func (dx *dclient) runStateSelecting() {
	dx.l.Printf("%s: Sending DHCPREQUEST for %s to %s\n", dx.iface.Name, dx.lastMsg.YourIP, dx.lastMsg.NextIP)

	xid := rand.Uint32()
	tmpl := msgtmpl.New(dx.iface, xid)
	rq := func() []byte { return tmpl.RequestSelecting(dx.lastMsg.YourIP, dx.lastMsg.NextIP) }
	var pass bool
	dx.lastMsg, dx.lastOpts, pass = dx.advanceState(verifySelectingAck(dx.lastMsg, xid), rq)
	if pass {
		dx.state = stateIfconfig
	} else {
		dx.state = stateInit
	}
}

func (dx *dclient) runStateIfconfig() {
	dx.l.Printf("%s: Configuring interface to use IP %s\n", dx.iface.Name, dx.lastMsg.YourIP)
	if err := libif.SetIface(dx.currentNetconfig()); err != nil {
		dx.l.Printf("%s: Unexpected error while configuring interface, falling back to INIT in 30 sec! (error was: %v)\n", dx.iface.Name, err)
		dx.state = stateInitIface

		// Sleep with context to not block the whole task.
		fctx, _ := context.WithTimeout(dx.ctx, time.Second*30)
		<-fctx.Done()
	} else {
		dx.state = stateBound
	}
}
