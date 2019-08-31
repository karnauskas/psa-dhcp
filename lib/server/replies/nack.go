package replies

import (
	"net"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
)

func AssembleNACK(xid uint32, srcIP net.IP, dstMAC net.HardwareAddr) []byte {
	return assembleUdp(srcIP, net.IPv4(0, 0, 0, 0), dhcpmsg.Message{
		Op:        dhcpmsg.OpReply,
		Xid:       xid,
		Htype:     dhcpmsg.HtypeETHER,
		Hops:      1,
		ClientMAC: dstMAC,
		Cookie:    dhcpmsg.DHCPCookie,
		Options: []dhcpmsg.DHCPOpt{
			dhcpmsg.OptionType(dhcpmsg.MsgTypeNack),
			dhcpmsg.OptionServerIdentifier(srcIP),
		},
	}.Assemble())
}
