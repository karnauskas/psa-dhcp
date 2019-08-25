package dclient

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/arpping"
	vy "gitlab.com/adrian_blx/psa-dhcp/lib/client/verify"
	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
	"gitlab.com/adrian_blx/psa-dhcp/lib/layer"
	"gitlab.com/adrian_blx/psa-dhcp/lib/rsocks"
)

var (
	errWasNack = fmt.Errorf("DHCP NACK received")
)

type senderFunc func() ([]byte, net.IP, net.IP)
type vrfyFunc func(dhcpmsg.Message, dhcpmsg.DecodedOptions) vy.State

type ssock interface {
	Close() error
	Write([]byte) (int, error)
}

// sendSock returns a multicast or unicast sending socket, depending on the to be sent contents.
func sendSocket(ctx context.Context, iface *net.Interface, sender senderFunc) (ssock, error) {
	_, src, dst := sender()
	if src != nil && dst != nil {
		for i := 0; i < 5 && ctx.Err() == nil; i++ {
			if hwaddr, err := arpping.Ping(ctx, iface, src, dst); err == nil {
				return rsocks.GetUnicastSendSock(iface, hwaddr)
			}
		}
	}
	return rsocks.GetIPSendSock(iface)
}

// sendMessage invokes the supplied 'sendFunc' function to send a message on the selected interface.
func sendMessage(ctx context.Context, iface *net.Interface, sender senderFunc) error {
	s, err := sendSocket(ctx, iface, sender)
	if err != nil {
		return err
	}
	defer s.Close()

	barrier := time.Second * 100
	delay := time.Millisecond * 700
	for {
		b, _, _ := sender()
		if _, err := s.Write(b); err != nil {
			return err
		}
		if delay < barrier {
			delay += time.Duration(rand.Int63() % (1 + delay.Nanoseconds()))
		}
		select {
		case <-time.After(delay):
			continue
		case <-ctx.Done():
			return nil
		}
	}
}

// catchReply returns a reply from a message triggered by sendMessage.
// Either an error is returned or a message which passed the verification function.
func catchReply(octx context.Context, iface *net.Interface, vrfy vrfyFunc) (dhcpmsg.Message, dhcpmsg.DecodedOptions, error) {
	s, err := rsocks.GetIPRecvSock(iface)
	if err != nil {
		return dhcpmsg.Message{}, dhcpmsg.DecodedOptions{}, err
	}
	ctx, cancel := context.WithCancel(octx)
	defer cancel()

	go func() {
		<-ctx.Done()
		s.Close()
	}()

	buff := make([]byte, 4096)
	for {
		nr, err := s.Read(buff)
		if err != nil {
			return dhcpmsg.Message{}, dhcpmsg.DecodedOptions{}, err
		}
		v4, err := layer.DecodeIPv4(buff[0:nr])
		if err != nil {
			continue
		}
		if v4.Protocol == 0x11 {
			if udp, err := layer.DecodeUDP(v4.Data); err == nil && udp.DstPort == 68 {
				if msg, err := dhcpmsg.Decode(udp.Data); err == nil && bytes.Equal(msg.ClientMAC[:], iface.HardwareAddr) {
					// Fixme: we shouldn't check the client mac from the dhcp payload here, that should be done/filtered on the receiving sock via BPF.
					opts := dhcpmsg.DecodeOptions(msg.Options)
					switch vrfy(*msg, opts) {
					case vy.Passed:
						return *msg, opts, nil
					case vy.IsNack:
						return *msg, opts, errWasNack
					}
				}
			}
		}
	}
}
