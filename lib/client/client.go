package client

import (
	"context"
	"log"
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/layer"
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

	zz := make([]byte, 272)
	uu := layer.UDP{
		SrcPort: 68,
		DstPort: 67,
		Data:    zz,
	}
	xx := layer.IPv4{
		Identification: 43062,
		Destination:    net.IPv4(255, 255, 255, 255),
		TTL:            250,
		Protocol:       17,
		Data:           uu.Assemble(),
	}
	for {
		s.Write(xx.Assemble())
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
