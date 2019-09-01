package replies

import (
	"net"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
)

func AssembleNACK(xid uint32, srcIP net.IP, dstMAC net.HardwareAddr) []byte {
	return assembleUdp(srcIP, net.IPv4bcast, dhcpmsg.Message{
		Op:        dhcpmsg.OpReply,
		Xid:       xid,
		Htype:     dhcpmsg.HtypeETHER,
		ClientMAC: dstMAC,
		Cookie:    dhcpmsg.DHCPCookie,
		Options: []dhcpmsg.DHCPOpt{
			dhcpmsg.OptionType(dhcpmsg.MsgTypeNack),
			dhcpmsg.OptionServerIdentifier(srcIP),
		},
	}.Assemble())
}
