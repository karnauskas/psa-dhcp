package client

import (
	"context"
	"log"
	"net"

	"gitlab.com/adrian_blx/psa-dhcp/lib/libif"
)

func Run(ctx context.Context, l *log.Logger, iface *net.Interface) error {
	l.Printf("Unconfiguring interface '%s'\n", iface.Name)
	if err := reInitIface(iface); err != nil {
		return err
	}
	return nil
}

func reInitIface(iface *net.Interface) error {
	if err := libif.Down(iface); err != nil {
		return err
	}
	if err := libif.Unconfigure(iface); err != nil {
		return err
	}
	if err := libif.Up(iface); err != nil {
		return err
	}
	return nil
}
