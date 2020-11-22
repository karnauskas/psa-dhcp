package replies

import (
	"net"

	"git.sr.ht/~adrian-blx/psa-dhcp/lib/dhcpmsg"
)

func AssembleACK(xid uint32, flags uint16, srcIP, dstIP net.IP, dstMAC net.HardwareAddr, opts []dhcpmsg.DHCPOpt) []byte {
	return assembleUdp(srcIP, dstFromFlag(flags, dstIP), dhcpmsg.Message{
		Op:        dhcpmsg.OpReply,
		YourIP:    dstIP,
		Xid:       xid,
		Flags:     flags,
		Htype:     dhcpmsg.HtypeETHER,
		ClientMAC: dstMAC,
		Cookie:    dhcpmsg.DHCPCookie,
		Options: append([]dhcpmsg.DHCPOpt{
			dhcpmsg.OptionType(dhcpmsg.MsgTypeAck),
			dhcpmsg.OptionServerIdentifier(srcIP),
		}, opts...),
	}.Assemble())
}
