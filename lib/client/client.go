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
	// Unconfigures the interface and brings it up.
	stateInitIface = iota
	// Send initial DHCPDISCOVER.
	stateInit
	// Selects a dhcp server via DHCPREQUEST.
	stateSelecting
	// Configures the OS with the received configuration.
	stateIfconfig
	// We have a lease and are sleeping.
	stateBound
	// We try to renew the state (unicast)
	stateRenewing
	// We try to rebind (broadcast)
	stateRebinding
)

const (
	minLeaseDuration = 10 * time.Second
	maxLeaseDuration = 30 * time.Minute
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
		xid := rand.Uint32()
		dx.l.Printf("In state %d\n", dx.state)
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
			// fixme: verify if the times make sense (x > a)
			now := time.Now()
			dx.boundDeadlines = boundDeadlines{
				t1: now.Add(normalizeLease(dx.lastOpts.RenewalTime)),
				t2: now.Add(normalizeLease(dx.lastOpts.RebindTime)),
				tx: now.Add(normalizeLease(dx.lastOpts.IPAddressLeaseTime)),
			}
			dx.l.Printf("Sleeping, deadlines are %+v\n", dx.boundDeadlines)
			hackAbsoluteSleep(dx.ctx, dx.boundDeadlines.t1)
			dx.state = stateRenewing
		case stateRenewing:
			dx.l.Printf("In state renewing until %s\n", dx.boundDeadlines.t2)
			// FIXME: This should be unicast with the correct mac.
			tmpl := msgtmpl.New(dx.iface, xid)
			rq := func() []byte { return tmpl.RequestRenewing(dx.lastMsg.YourIP, dx.lastOpts.ServerIdentifier) }

			if lm, lo, p := dx.advanceState(dx.boundDeadlines.t2, verifyRenewAck(dx.lastMsg, xid), rq); p {
				dx.state = stateIfconfig
				dx.lastMsg = lm
				dx.lastOpts = lo
			} else {
				dx.state = stateRebinding
			}
		case stateRebinding:
			// fixme: this must not use the target IP.
			dx.l.Printf("In state rebinding until %s\n", dx.boundDeadlines.tx)
			tmpl := msgtmpl.New(dx.iface, xid)
			rq := func() []byte { return tmpl.RequestRebinding(dx.lastMsg.YourIP) }
			if lm, lo, p := dx.advanceState(dx.boundDeadlines.tx, verifyRebindingAck(dx.lastMsg, xid), rq); p {
				dx.state = stateIfconfig
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
		Interface:     dx.iface,
		Router:        dx.lastOpts.Routers[0],
		IP:            dx.lastMsg.YourIP,
		Cidr:          cidr,
		LeaseDuration: normalizeLease(dx.lastOpts.RebindTime),
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

	dx.l.Printf("advanceState until %s\n", deadline)

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

	xid := rand.Uint32()
	tmpl := msgtmpl.New(dx.iface, xid)
	if lm, lo, p := dx.advanceState(time.Now().Add(time.Minute), verifyOffer(xid), func() []byte { return tmpl.Discover() }); p {
		dx.state = stateSelecting
		dx.lastMsg = lm
		dx.lastOpts = lo
	}
	// else: can't advance to any other state.
}

// runStateSelecting selects a dhcp server by *broadcasting* a DHCPREQUEST.
func (dx *dclient) runStateSelecting() {
	dx.l.Printf("Sending DHCPREQUEST for %s to %s\n", dx.lastMsg.YourIP, dx.lastMsg.NextIP)

	xid := rand.Uint32()
	tmpl := msgtmpl.New(dx.iface, xid)
	rq := func() []byte { return tmpl.RequestSelecting(dx.lastMsg.YourIP, dx.lastMsg.NextIP) }
	if lm, lo, p := dx.advanceState(time.Now().Add(time.Minute), verifySelectingAck(dx.lastMsg, xid), rq); p {
		dx.state = stateIfconfig
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
		dx.l.Printf("Unexpected error while configuring interface, falling back to INIT in 30 sec! (error was: %v)\n", err)
		dx.state = stateInitIface

		// Sleep with context to not block the whole task.
		fctx, _ := context.WithTimeout(dx.ctx, time.Second*30)
		<-fctx.Done()
	} else {
		dx.state = stateBound
	}
}
