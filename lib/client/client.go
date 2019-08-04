package client

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

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
			udp, err := layer.DecodeUDP(v4.Data)
			_ = err
			fmt.Printf("RAW = %+v\n", buff[0:nr])
			fmt.Printf("V4 = %+v\n", v4)
			fmt.Printf("UDP= %+v\n\n", udp)
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
	fmt.Printf(">> %T\n", iface.HardwareAddr)

	var mac [6]byte
	copy(mac[:], iface.HardwareAddr[0:len(mac)])
	for {
		zz := dhcpmsg.Message{
			Op:        1,
			Htype:     1,
			Hlen:      uint8(len(mac)),
			Hops:      0,
			Xid:       rand.Uint32(),
			Secs:      0,
			Flags:     dhcpmsg.FlagBroadcast,
			ClientMAC: mac,
			Cookie:    dhcpmsg.DHCPCookie,
			Options: []dhcpmsg.DHCPOpt{
				dhcpmsg.OptDiscover(),
				dhcpmsg.OptHostname("abyssloch"),
			},
		}.Assemble()
		xx := layer.IPv4{
			Identification: uint16(rand.Uint32()),
			Destination:    net.IPv4(255, 255, 255, 255),
			Source:         net.IPv4(0, 0, 0, 0),
			TTL:            250,
			Protocol:       layer.ProtoUDP,
			Data: layer.UDP{
				SrcPort: 68,
				DstPort: 67,
				Data:    zz}.Assemble(),
		}.Assemble()
		s.Write(xx)
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
