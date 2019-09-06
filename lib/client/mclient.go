package client

import (
	"context"
	"log"
	"net"

	cb "gitlab.com/adrian_blx/psa-dhcp/lib/client/callback"
	"gitlab.com/adrian_blx/psa-dhcp/lib/client/dclient"
	"gitlab.com/adrian_blx/psa-dhcp/lib/ifmon"
)

// mclient is the 'main' client and is what we return on New.
type mclient struct {
	ctx    context.Context
	l      *log.Logger
	iface  *net.Interface
	script string
}

// New returns a new mclient to the caller. Use Run() to launch it.
func New(ctx context.Context, l *log.Logger, iface *net.Interface, script string) *mclient {
	return &mclient{ctx: ctx, l: l, iface: iface, script: script}
}

// Run runs the main loop.
func (mx *mclient) Run() error {
	// This is the context we use for the dclient.
	// We will cancel it if there are any important netlink changes.
	dctx, dcancel := context.WithCancel(mx.ctx)
	defer dcancel()

	// Check for interface changes, this will trigger dcancel.
	// We use a pointer as the local value will get updated.
	go mx.monitor(&dcancel)

	dx := dclient.New(dctx, mx.iface, mx.l, mx.filterNetconfig, cb.Cbhandler(mx.script, mx.iface, mx.l))
	for {
		dx.Run()
		if err := mx.ctx.Err(); err != nil {
			return err
		}
		dctx, dcancel = context.WithCancel(mx.ctx)
		dx.ResumeClient(dctx)
	}
}

// monitor checks the interface for important changes.
func (mx *mclient) monitor(cancel *context.CancelFunc) {
	ev := make(chan bool, 64)
	defer close(ev)
	go ifmon.MonitorChanges(mx.ctx, mx.iface, ev)
	for {
		select {
		case <-ev:
			mx.l.Printf("Interface %s is now 'up'", mx.iface.Name)
			flushChan(ev)
			(*cancel)()
		case <-mx.ctx.Done():
			return
		}
	}

}

// flushChan consumes all messages on a channel.
func flushChan(c chan bool) {
	for {
		select {
		case <-c:
			// just eat it!
		default:
			return
		}
	}
}
