package replies

import (
	"net"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
)

func AssembleOffer(xid uint32, flags uint16, srcIP, dstIP net.IP, dstMAC net.HardwareAddr, opts []dhcpmsg.DHCPOpt) []byte {
	return assembleUdp(srcIP, dstFromFlag(flags, dstIP), dhcpmsg.Message{
		Op:        dhcpmsg.OpReply,
		Xid:       xid,
		Htype:     dhcpmsg.HtypeETHER,
		YourIP:    dstIP,
		Flags:     flags,
		ClientMAC: dstMAC,
		Cookie:    dhcpmsg.DHCPCookie,
		Options: append([]dhcpmsg.DHCPOpt{
			dhcpmsg.OptionType(dhcpmsg.MsgTypeOffer),
			dhcpmsg.OptionServerIdentifier(srcIP),
		}, opts...),
	}.Assemble())
}
