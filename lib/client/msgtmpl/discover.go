package msgtmpl

import (
	"math/rand"
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
	"gitlab.com/adrian_blx/psa-dhcp/lib/layer"
)

func (rx *tmpl) Discover() []byte {
	rx.lastSecs = uint16(time.Now().Sub(rx.start).Seconds())
	pl := layer.IPv4{
		Identification: uint16(rand.Uint32()),
		Destination:    net.IPv4(255, 255, 255, 255),
		Source:         net.IPv4(0, 0, 0, 0),
		TTL:            250,
		Protocol:       layer.ProtoUDP,
		Data: layer.UDP{
			SrcPort: 68,
			DstPort: 67,
			Data: dhcpmsg.Message{
				Op:        dhcpmsg.OpRequest,
				Htype:     dhcpmsg.HtypeIEEE802,
				Hlen:      uint8(len(rx.hwaddr)),
				Xid:       rx.xid,
				Secs:      rx.lastSecs,
				Flags:     dhcpmsg.FlagBroadcast, // We always send to 255.255.255.255.
				ClientMAC: rx.hwaddr,
				Cookie:    dhcpmsg.DHCPCookie,
				Options: []dhcpmsg.DHCPOpt{
					dhcpmsg.OptionType(dhcpmsg.MsgTypeDiscover),
					dhcpmsg.OptionHostname(rx.hostname),
				},
			}.Assemble(),
		}.Assemble(),
	}.Assemble()
	return pl
}
