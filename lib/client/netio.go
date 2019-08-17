package client

import (
	"bytes"
	"context"
	"math/rand"
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/client/arpping"
	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
	"gitlab.com/adrian_blx/psa-dhcp/lib/layer"
	"gitlab.com/adrian_blx/psa-dhcp/lib/rsocks"
)

type senderFunc func() ([]byte, net.IP, net.IP)

type ssock interface {
	Close() error
	Write([]byte) (int, error)
}

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

	barrier := time.Second * 45
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

// catchReply waits for DHCP messages and returns them using the supplied channel.
// Returning a 'nil' message indicates that this function returned.
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
					if vrfy(*msg, opts) {
						return *msg, opts, nil
					}
				}
			}
		}
	}
}
