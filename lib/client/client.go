package client

import (
	"context"
	"log"
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
	s, err := rsocks.GetRawSendSock(iface)
	if err != nil {
		return err
	}

	zz := dhcpmsg.Message{
		Op:        1,
		Htype:     1,
		Hlen:      6,
		Hops:      0,
		Xid:       0xb4db4b3,
		Secs:      0,
		Flags:     dhcpmsg.FlagBroadcast,
		ClientMAC: [6]byte{0xf4, 0x8c, 0x50, 0xe8, 0xdf, 0x32},
		Options: []dhcpmsg.DHCPOpt{
			dhcpmsg.OptDiscover(),
			dhcpmsg.OptHostname("abyssloch"),
		},
	}.Assemble()
	uu := layer.UDP{
		SrcPort: 68,
		DstPort: 67,
		Data:    zz,
	}.Assemble()
	xx := layer.IPv4{
		Identification: 26174,
		Destination:    net.IPv4(255, 255, 255, 255),
		Source:         net.IPv4(0, 0, 0, 0),
		TTL:            250,
		Protocol:       17,
		Data:           uu,
	}.Assemble()
	for {
		s.Write(xx)
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
