package replies

import (
	"net"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
)

func AssembleACK(xid uint32, srcIP, dstIP net.IP, dstMAC net.HardwareAddr, opts []dhcpmsg.DHCPOpt) []byte {
	return assembleUdp(srcIP, dstIP, dhcpmsg.Message{
		Op:        dhcpmsg.OpReply,
		YourIP:    dstIP,
		Xid:       xid,
		Htype:     dhcpmsg.HtypeETHER,
		ClientMAC: dstMAC,
		Cookie:    dhcpmsg.DHCPCookie,
		Options: append([]dhcpmsg.DHCPOpt{
			dhcpmsg.OptionType(dhcpmsg.MsgTypeAck),
			dhcpmsg.OptionServerIdentifier(srcIP),
		}, opts...),
	}.Assemble())
}
