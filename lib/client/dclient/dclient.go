package dclient

import (
	"context"
	"log"
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
	"gitlab.com/adrian_blx/psa-dhcp/lib/libif"
	"golang.org/x/time/rate"
)

const (
	stateINVALID_       = iota
	statePurgeInterface // Unconfigures the interface and brings it up.
	stateDiscovering    // Send initial DHCPDISCOVER.
	stateSelecting      // Selects a dhcp server via DHCPREQUEST.
	stateArpCheck       // Verify the current config by sending an ARP ping
	stateIfconfig       // Configures the OS with the received configuration.
	stateBound          // We have a lease and are sleeping.
	stateRenewing       // We try to renew the state (unicast)
	stateRebinding      // We try to rebind (broadcast)

)

type boundDeadlines struct {
	t1 time.Time // When to enter renewing state
	t2 time.Time // When to enter rebinding state
	tx time.Time // Total lease time
}

// dclient is the dhcp client state machine.
type dclient struct {
	ctx            context.Context                          // The context to use.
	l              *log.Logger                              // Logging interface
	iface          *net.Interface                           // Network hardware interface
	state          int                                      // The current state we are in
	lastMsg        dhcpmsg.Message                          // Last accepted DHCP reply
	lastOpts       dhcpmsg.DecodedOptions                   // Options of last accepted reply
	boundDeadlines boundDeadlines                           // Deadline information, updated by BOUND state
	limiter        *rate.Limiter                            // Rate limiter
	callbacks      []func(context.Context, *libif.Ifconfig) // List of possible callbacks.
}

func New(ctx context.Context, iface *net.Interface, l *log.Logger, cb []func(context.Context, *libif.Ifconfig)) *dclient {
	return &dclient{
		ctx:       ctx,
		iface:     iface,
		l:         l,
		state:     statePurgeInterface,
		callbacks: cb,
		limiter:   rate.NewLimiter(1, 10),
	}
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

		if !dx.limiter.Allow() {
			dx.l.Panicf("client went bananas, consumed all tokens!")
		}
		// break if main context is done.
		if err := dx.ctx.Err(); err != nil {
			return err
		}
	}
}

// resumeClient re-inits a client using a new context.
func (dx *dclient) ResumeClient(ctx context.Context) {
	if dx.state == stateBound || dx.state == stateRenewing || dx.state == stateRebinding {
		// If we are in any of these states, we try to quickly re-validate our config by jumping into the rebinding state.
		// This might enable us to confirm our lease and jump back into a bound state without bringing the interface fully down.
		sd := time.Now().Add(5 * time.Second)
		dx.boundDeadlines = boundDeadlines{
			t1: sd,
			t2: sd,
			tx: sd,
		}
		dx.state = stateRebinding
	} else {
		dx.state = statePurgeInterface
	}
	dx.ctx = ctx
}

func (dx dclient) buildNetconfig() libif.Ifconfig {
	netmask := dx.lastMsg.YourIP.DefaultMask()
	if _, bits := dx.lastOpts.SubnetMask.Size(); bits != 0 {
		netmask = dx.lastOpts.SubnetMask
	}

	c := libif.Ifconfig{
		Interface:     dx.iface,
		MTU:           int(dx.lastOpts.InterfaceMTU),
		Router:        dx.lastOpts.Routers[0],
		IP:            dx.lastMsg.YourIP,
		Netmask:       netmask,
		DNS:           dx.lastOpts.DNS,
		DomainName:    dx.lastOpts.DomainName,
		LeaseDuration: dx.lastOpts.IPAddressLeaseDuration,
	}
	return c
}

// hackAbsoluteSleep sleeps until the context expires or the supplied time was reached.
// Go does not expose access to TIMER_ABSTIME, so we poll once in a while to detect clock
// jumps due to hyberntion. (Which in reality most likely also cause netlink events).
func hackAbsoluteSleep(ctx context.Context, when time.Time) {
	for {
		left := when.Sub(time.Now())
		select {
		case <-ctx.Done():
			return
		case <-time.After(left):
			return
		case <-time.After(17 * time.Second):
			// -> recalculate time
		}
	}
}

func (dx *dclient) advanceState(deadline time.Time, vrfy vrfyFunc, sender senderFunc) (dhcpmsg.Message, dhcpmsg.DecodedOptions, error) {
	ctx, cancel := context.WithDeadline(dx.ctx, deadline)
	defer cancel()

	dx.l.Printf("  ==> waiting for valid reply until %s", deadline.Format(time.RFC3339))
	go sendMessage(ctx, dx.iface, sender)
	msg, opts, err := catchReply(ctx, dx.iface, vrfy)

	if err != nil {
		// If there was an error, wait until the context expires (if we might have a
		// sock setup error) to avoid flooding the line.
		<-ctx.Done()
		return msg, opts, err
	}
	return msg, opts, nil
}

func (dx *dclient) runCallbacks(c *libif.Ifconfig) {
	for _, cb := range dx.callbacks {
		cb(dx.ctx, c)
	}
}
