package client

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/client/msgtmpl"
	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
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
	go catchDiscover(iface)
	return sendDiscover(iface)
}

func catchDiscover(iface *net.Interface) error {
	s, err := rsocks.GetRawRecvSock(iface)
	if err != nil {
		return err
	}
	buff := make([]byte, 1024)
	for {
		nr, err := s.Read(buff)
		v4, err := layer.DecodeIPv4(buff[0:nr])
		if err != nil {
			fmt.Printf("V4 err: %v\n", err)
			continue
		}
		if v4.Protocol == 0x11 {
			if udp, err := layer.DecodeUDP(v4.Data); err == nil {
				msg, err := dhcpmsg.Decode(udp.Data)
				fmt.Printf("%v -> %+v\n", err, msg)
			}
		}
	}
	s.Close()
	return nil
}

func sendDiscover(iface *net.Interface) error {
	s, err := rsocks.GetRawSendSock(iface)
	if err != nil {
		return err
	}

	tmpl := msgtmpl.New(iface)
	for {
		s.Write(tmpl.Discover())
		//s.SendDiscover()
		time.Sleep(time.Second * 5)
	}
	s.Close()
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
