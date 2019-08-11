package msgtmpl

import (
	"math/rand"
	"net"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
	"gitlab.com/adrian_blx/psa-dhcp/lib/layer"
)

func (rx *tmpl) RequestSelecting(requestedIP, serverIdentifier net.IP) []byte {
	return rx.request(net.IPv4(0, 0, 0, 0), net.IPv4(255, 255, 255, 255),
		&requestedIP, &serverIdentifier)
}

func (rx *tmpl) request(sourceIP, destinationIP net.IP, requestedIP, serverIdentifier *net.IP) []byte {
	msgopts := []dhcpmsg.DHCPOpt{
		dhcpmsg.OptionType(dhcpmsg.MsgTypeRequest),
		dhcpmsg.OptionHostname(rx.hostname),
	}
	if requestedIP != nil {
		msgopts = append(msgopts, dhcpmsg.OptionRequestedIP(*requestedIP))
	}
	if serverIdentifier != nil {
		msgopts = append(msgopts, dhcpmsg.OptionServerIdentifier(*serverIdentifier))
	}

	pl := layer.IPv4{
		Identification: uint16(rand.Uint32()),
		Destination:    destinationIP,
		Source:         sourceIP,
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
				ClientIP:  sourceIP,
				Cookie:    dhcpmsg.DHCPCookie,
				Options:   msgopts,
			}.Assemble(),
		}.Assemble(),
	}.Assemble()
	return pl
}
