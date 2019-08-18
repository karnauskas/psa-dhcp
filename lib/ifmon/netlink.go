package ifmon

import (
	"context"
	"net"
	"syscall"

	"github.com/vishvananda/netlink"
)

// MonitorChanges checks for when the given interface came 'back up'.
func MonitorChanges(ctx context.Context, iface *net.Interface, event chan<- bool) error {
	update := make(chan netlink.LinkUpdate)
	done := make(chan struct{})

	if err := netlink.LinkSubscribe(update, done); err != nil {
		return err
	}
	defer close(done)

	oldState := uint32(syscall.IFF_RUNNING)
	for {
		select {
		case msg := <-update:
			flags := msg.IfInfomsg.Flags & syscall.IFF_RUNNING
			if flags != oldState && iface.Index == int(msg.IfInfomsg.Index) {
				oldState = flags
				if flags&syscall.IFF_RUNNING != 0 {
					event <- true
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}
