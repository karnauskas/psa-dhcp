package client

import (
	"context"
	"log"
	"net"

	cb "git.sr.ht/~adrian-blx/psa-dhcp/lib/client/callback"
	"git.sr.ht/~adrian-blx/psa-dhcp/lib/client/dclient"
	"git.sr.ht/~adrian-blx/psa-dhcp/lib/ifmon"
)

// mclient is the 'main' client and is what we return on New.
type mclient struct {
	l              *log.Logger
	iface          *net.Interface
	script         string
	configureRoute bool
}

// New returns a new mclient to the caller. Use Run() to launch it.
func New(l *log.Logger, iface *net.Interface, script string, croute bool) *mclient {
	return &mclient{
		l:              l,
		iface:          iface,
		script:         script,
		configureRoute: croute,
	}
}

// Run runs the main loop.
func (mx *mclient) Run(ctx context.Context) error {
	// This is the context we use for the dclient.
	// We will cancel it if there are any important netlink changes.
	dctx, dcancel := context.WithCancel(ctx)
	defer dcancel()

	// Check for interface changes, this will trigger dcancel.
	// We use a pointer as the local value will get updated.
	go mx.monitor(ctx, &dcancel)

	dx := dclient.New(dctx, mx.iface, mx.l, mx.filterNetconfig, cb.Cbhandler(mx.script, mx.iface, mx.l))
	for {
		dx.Run()
		if err := ctx.Err(); err != nil {
			return err
		}
		dctx, dcancel = context.WithCancel(ctx)
		dx.ResumeClient(dctx)
	}
}

// monitor checks the interface for important changes.
func (mx *mclient) monitor(ctx context.Context, cancel *context.CancelFunc) {
	ev := make(chan bool, 64)
	defer close(ev)
	go ifmon.MonitorChanges(ctx, mx.iface, ev)
	for {
		select {
		case <-ev:
			mx.l.Printf("Interface %s is now 'up'", mx.iface.Name)
			flushChan(ev)
			(*cancel)()
		case <-ctx.Done():
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
