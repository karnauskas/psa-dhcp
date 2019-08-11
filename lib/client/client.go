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

const (
	stateInitIface = iota
	stateInit
	stateSelecting
	stateIfconfig
)

type dclient struct {
	ctx      context.Context
	l        *log.Logger
	iface    *net.Interface
	state    int
	lastMsg  dhcpmsg.Message
	lastOpts dhcpmsg.DecodedOptions
}

type vrfyFunc func(dhcpmsg.Message, dhcpmsg.DecodedOptions) bool
type senderFunc func() []byte

func New(ctx context.Context, l *log.Logger, iface *net.Interface) *dclient {
	return &dclient{ctx: ctx, l: l, iface: iface, state: stateInitIface}
}

func (dx *dclient) Run() error {
	var pass bool
	for {
		xid := rand.Uint32()

		switch dx.state {
		case stateInitIface:
			dx.l.Printf("Downing iface\n")
			reInitIface(dx.iface)
			dx.state = stateInit
		case stateInit:
			tmpl := msgtmpl.New(dx.iface, xid)
			dx.lastMsg, dx.lastOpts, pass = dx.advanceState(verifyDiscover(xid), func() []byte { return tmpl.Discover() })
			if pass {
				dx.state = stateSelecting
			}
		case stateSelecting:
			tmpl := msgtmpl.New(dx.iface, xid)
			rq := func() []byte { return tmpl.RequestSelecting(dx.lastMsg.YourIP, *dx.lastOpts.ServerIdentifier) }
			dx.lastMsg, dx.lastOpts, pass = dx.advanceState(verifyAck(dx.lastMsg, xid), rq)
			dx.state = 33
			if pass {
				dx.state = stateIfconfig
			} else {
				dx.state = stateInit
			}
		case stateIfconfig:
			fmt.Printf(">> SHOULD IFCONFIG MYSELF\n")
			libif.SetIface(dx.iface, dx.lastMsg.YourIP, (*dx.lastOpts.Routers)[0], dx.lastOpts.SubnetMask)
			dx.state = 99
		default:
			dx.l.Panicf("invalid state: %d\n", dx.state)
		}
	}
}

func (dx *dclient) advanceState(vrfy vrfyFunc, sender senderFunc) (reply dhcpmsg.Message, opts dhcpmsg.DecodedOptions, pass bool) {
	dctx, dcancel := context.WithTimeout(dx.ctx, time.Second*15)
	defer dcancel()

	c := make(chan *dhcpmsg.Message)
	go sendMessage(dctx, dx.iface, sender)
	go catchReply(dctx, dx.iface, c)

	// Message will be empty if we never got something or nil
	// when catcher exited.
	msg := &dhcpmsg.Message{}
xloop:
	for {
		select {
		case <-dctx.Done():
			break xloop
		case msg = <-c:
			if msg != nil {
				reply = *msg
				opts = dhcpmsg.DecodeOptions(reply.Options)
				if !vrfy(reply, opts) {
					dx.l.Printf("Received message did not pass verification\n")
					continue
				} else {
					pass = true
				}
			}
			dcancel()
			break xloop
		}
	}

	// receiving 'nil' indicates that the catcher was shutdown.
	for msg != nil {
		msg = <-c
	}
	close(c)
	return
}

func catchReply(ctx context.Context, iface *net.Interface, c chan *dhcpmsg.Message) {
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
				if msg, err := dhcpmsg.Decode(udp.Data); err == nil {
					c <- msg
				}
			}
		}
	}
}

func sendMessage(ctx context.Context, iface *net.Interface, sender func() []byte) error {
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
	var lerr error
	if err := libif.Down(iface); err != nil {
		lerr = err
	}
	if err := libif.Unconfigure(iface); err != nil {
		lerr = err
	}
	if err := libif.Up(iface); err != nil {
		lerr = err
	}
	return lerr
}
