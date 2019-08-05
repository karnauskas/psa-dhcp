package client

import (
	"context"
	"fmt"
	"log"
	"math/rand"
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

	dctx, dcancel := context.WithTimeout(ctx, time.Second*15)
	defer dcancel()

	xid := rand.Uint32()
	tmpl := msgtmpl.New(iface, xid)
	sender := func() []byte {
		return tmpl.Discover()
	}

	c := make(chan *dhcpmsg.Message)
	go sendDiscover(dctx, iface, sender)
	go catchDiscover(dctx, iface, xid, c)

	// Message will be empty if we never got something or nil
	// when catcher exited.
	msg := &dhcpmsg.Message{}
xloop:
	for {
		select {
		case <-dctx.Done():
			break xloop
		case msg = <-c:
			dcancel()
			break xloop
		}
	}

	reply := msg
	// receiving 'nil' indicates that the catcher was shutdown.
	for msg != nil {
		msg = <-c
	}
	close(c)

	fmt.Printf("Received reply: %+v\n", reply)
	if reply != nil {
		for _, o := range reply.Options {
			fmt.Printf("opt=%d => %+v ; %s\n", o.Option, o.Data, string(o.Data))
		}
	}
	return nil
}

func catchDiscover(ctx context.Context, iface *net.Interface, xid uint32, c chan *dhcpmsg.Message) {
	s, err := rsocks.GetRawRecvSock(iface)
	if err != nil {
		c <- nil
		return
	}
	defer s.Close()

	go func() {
		<-ctx.Done()
		s.Close()
	}()

	buff := make([]byte, 4096)
	for {
		nr, err := s.Read(buff)
		if err != nil {
			c <- nil
			return
		}
		v4, err := layer.DecodeIPv4(buff[0:nr])
		if err != nil {
			continue
		}
		if v4.Protocol == 0x11 {
			if udp, err := layer.DecodeUDP(v4.Data); err == nil && udp.DstPort == 68 {
				if msg, err := dhcpmsg.Decode(udp.Data); err == nil && msg.Xid == xid {
					c <- msg
				}
			}
		}
	}
}

func sendDiscover(ctx context.Context, iface *net.Interface, sender func() []byte) error {
	s, err := rsocks.GetRawSendSock(iface)
	if err != nil {
		return err
	}
	defer s.Close()

	for {
		if _, err := s.Write(sender()); err != nil {
			return err
		}
		select {
		case <-time.After(time.Second * 5):
			continue
		case <-ctx.Done():
			return nil
		}
	}
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
