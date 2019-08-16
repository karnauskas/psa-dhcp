package client

import (
	"context"
	"math/rand"
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
	"gitlab.com/adrian_blx/psa-dhcp/lib/layer"
	"gitlab.com/adrian_blx/psa-dhcp/lib/rsocks"
)

type senderFunc func() []byte

// sendMessage invokes the supplied 'sendFunc' function to send a message on the selected interface.
func sendMessage(ctx context.Context, iface *net.Interface, sender senderFunc) error {
	s, err := rsocks.GetIPSendSock(iface)
	if err != nil {
		return err
	}
	defer s.Close()

	barrier := time.Second * 45
	delay := time.Millisecond * 700
	for {
		if _, err := s.Write(sender()); err != nil {
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
				if msg, err := dhcpmsg.Decode(udp.Data); err == nil {
					opts := dhcpmsg.DecodeOptions(msg.Options)
					if vrfy(*msg, opts) {
						return *msg, opts, nil
					}
				}
			}
		}
	}
}