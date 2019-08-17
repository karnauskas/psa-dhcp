package client

import (
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/client/msgtmpl"
	vy "gitlab.com/adrian_blx/psa-dhcp/lib/client/verify"
)

// runStateDiscovering broadcasts a DHCPDISCOVER message on the interface.
func (dx *dclient) runStateDiscovering(nextState int) {
	dx.l.Printf("Sending DHCPDISCOVER broadcast\n")

	rq, xid := msgtmpl.Discover(dx.iface)
	if lm, lo, err := dx.advanceState(time.Now().Add(10*time.Minute), vy.VerifyOffer(xid), rq); err == nil {
		dx.lastMsg = lm
		dx.lastOpts = lo
		dx.state = nextState
	}
	// else: can't advance to any other state.
}

// runStateSelecting selects a dhcp server by *broadcasting* a DHCPREQUEST.
func (dx *dclient) runStateSelecting(nextState, failState int) {
	dx.l.Printf("Accepting offer for IP %s from server %s\n", dx.lastMsg.YourIP, dx.lastOpts.ServerIdentifier)

	rq, xid := msgtmpl.RequestSelecting(dx.iface, dx.lastMsg.YourIP, dx.lastOpts.ServerIdentifier)
	if lm, lo, err := dx.advanceState(time.Now().Add(time.Minute), vy.VerifySelectingAck(dx.lastMsg, dx.lastOpts, xid), rq); err == nil {
		dx.lastMsg = lm
		dx.lastOpts = lo
		dx.state = nextState
	} else {
		dx.state = failState
	}
}

func (dx *dclient) runStateBound(nextState int) {
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
	dx.l.Printf("-> Lease is valid for %s", dx.boundDeadlines.tx.Sub(now))
	dx.l.Printf("-> Renew will happen after %s, must rebind after %s", dx.boundDeadlines.t1.Sub(now), dx.boundDeadlines.t2.Sub(now))
	hackAbsoluteSleep(dx.ctx, dx.boundDeadlines.t1)
	dx.state = nextState
}

func (dx *dclient) runStateRenewing(nextState, failState int) {
	dx.l.Printf("Renewing lease, will try until %s", dx.boundDeadlines.t2)
	rq, xid := msgtmpl.RequestRenewing(dx.iface, dx.lastMsg.YourIP, dx.lastOpts.ServerIdentifier)
	if lm, lo, err := dx.advanceState(dx.boundDeadlines.t2, vy.VerifyRenewingAck(dx.lastMsg, dx.lastOpts, xid), rq); err == nil {
		dx.lastMsg = lm
		dx.lastOpts = lo
		dx.state = nextState
	} else {
		dx.state = failState
	}
}

func (dx *dclient) runStateRebinding(nextState, failState int) {
	dx.l.Printf("Rebindign lease, will try until %s", dx.boundDeadlines.tx)
	rq, xid := msgtmpl.RequestRebinding(dx.iface, dx.lastMsg.YourIP)
	if lm, lo, err := dx.advanceState(dx.boundDeadlines.tx, vy.VerifyRebindingAck(dx.lastMsg, dx.lastOpts, xid), rq); err == nil {
		dx.lastMsg = lm
		dx.lastOpts = lo
		dx.state = nextState
	} else {
		dx.state = failState
	}
}
