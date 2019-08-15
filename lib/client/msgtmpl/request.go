package msgtmpl

import (
	"math/rand"
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
	"gitlab.com/adrian_blx/psa-dhcp/lib/layer"
)

var (
	ipNone     = net.IPv4(0, 0, 0, 0)
	ipBcast    = net.IPv4(255, 255, 255, 255)
	maxMsgSize = uint16(1500)
)

func (rx *tmpl) request(msgtype uint8, sourceIP, destinationIP net.IP, requestedIP, serverIdentifier *net.IP) []byte {
	msgopts := []dhcpmsg.DHCPOpt{
		dhcpmsg.OptionType(msgtype),
		dhcpmsg.OptionClientIdentifier(rx.hwaddr),
		dhcpmsg.OptionMaxMessageSize(maxMsgSize),
	}
	if rx.hostname != "" {
		msgopts = append(msgopts, dhcpmsg.OptionHostname(rx.hostname))
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
				Htype:     dhcpmsg.HtypeETHER,
				Hlen:      uint8(len(rx.hwaddr)),
				Xid:       rx.xid,
				Secs:      uint16(time.Now().Sub(rx.start).Seconds()),
				ClientMAC: rx.hwaddr,
				ClientIP:  sourceIP,
				Cookie:    dhcpmsg.DHCPCookie,
				Options:   msgopts,
			}.Assemble(),
		}.Assemble(),
	}.Assemble()
	return pl
}
