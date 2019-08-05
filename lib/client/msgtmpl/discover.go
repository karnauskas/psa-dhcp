package msgtmpl

import (
	"math/rand"
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
	"gitlab.com/adrian_blx/psa-dhcp/lib/layer"
)

func (rx tmpl) Discover() []byte {
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
				Secs:      uint16(time.Now().Sub(rx.start).Seconds()),
				Flags:     dhcpmsg.FlagBroadcast,
				ClientMAC: rx.hwaddr,
				Cookie:    dhcpmsg.DHCPCookie,
				Options: []dhcpmsg.DHCPOpt{
					dhcpmsg.OptionDiscover(),
					dhcpmsg.OptionHostname(rx.hostname),
				},
			}.Assemble(),
		}.Assemble(),
	}.Assemble()
	return pl
}
