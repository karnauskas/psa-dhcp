package msgtmpl

import (
	"math/rand"
	"net"

	"git.sr.ht/~adrian-blx/psa-dhcp/lib/dhcpmsg"
	"git.sr.ht/~adrian-blx/psa-dhcp/lib/layer"
)

var (
	ipNone     = net.IPv4(0, 0, 0, 0)
	ipBcast    = net.IPv4(255, 255, 255, 255)
	maxMsgSize = uint16(1500)
)

func (rx *tmpl) request(msgtype uint8, sourceIP, destinationIP, requestedIP, serverIdentifier net.IP) []byte {
	msgopts := []dhcpmsg.DHCPOpt{
		dhcpmsg.OptionType(msgtype),
		dhcpmsg.OptionClientIdentifier(rx.hwaddr),
		dhcpmsg.OptionMaxMessageSize(maxMsgSize),
		dhcpmsg.OptionParametersList(
			dhcpmsg.OptSubnetMask, dhcpmsg.OptRouter, dhcpmsg.OptIPAddressLeaseDuration,
			dhcpmsg.OptServerIdentifier,
			dhcpmsg.OptDNS, dhcpmsg.OptDomainName, dhcpmsg.OptInterfaceMTU,
			dhcpmsg.OptRenewalDuration, dhcpmsg.OptRebindDuration),
	}
	if requestedIP != nil {
		msgopts = append(msgopts, dhcpmsg.OptionRequestedIP(requestedIP))
	}
	if serverIdentifier != nil {
		msgopts = append(msgopts, dhcpmsg.OptionServerIdentifier(serverIdentifier))
	}

	pl := layer.IPv4{
		Identification: uint16(rand.Uint32()),
		Destination:    destinationIP,
		Source:         sourceIP,
		TTL:            64,
		Protocol:       layer.ProtoUDP,
		Data: layer.UDP{
			SrcPort: 68,
			DstPort: 67,
			Data: dhcpmsg.Message{
				Op:        dhcpmsg.OpRequest,
				Htype:     dhcpmsg.HtypeETHER,
				Xid:       rx.xid,
				ClientMAC: rx.hwaddr,
				ClientIP:  sourceIP,
				Cookie:    dhcpmsg.DHCPCookie,
				Options:   msgopts,
			}.Assemble(),
		}.Assemble(),
	}.Assemble()
	return pl
}
