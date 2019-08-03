package client

import (
	"context"
	"log"
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/libif"
	"gitlab.com/adrian_blx/psa-dhcp/lib/rsocks"
)

func Run(ctx context.Context, l *log.Logger, iface *net.Interface) error {
	l.Printf("Unconfiguring interface '%s'\n", iface.Name)
	/*	if err := reInitIface(iface); err != nil {
			return err
		}
	*/
	s, err := rsocks.GetRawSendSock(iface)
	if err != nil {
		return err
	}
	for {
		s.Write([]byte{1, 2, 3})
		//s.SendDiscover()
		time.Sleep(time.Second * 5)
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
