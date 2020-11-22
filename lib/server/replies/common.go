package replies

import (
	"net"

	"git.sr.ht/~adrian-blx/psa-dhcp/lib/dhcpmsg"
	"git.sr.ht/~adrian-blx/psa-dhcp/lib/layer"
)

func assembleUdp(srcIP, dstIP net.IP, payload []byte) []byte {
	return layer.IPv4{
		TTL:         64,
		Protocol:    0x11,
		Source:      srcIP,
		Destination: dstIP,
		Data: layer.UDP{
			SrcPort: 67,
			DstPort: 68,
			Data:    payload,
		}.Assemble(),
	}.Assemble()
}

func dstFromFlag(flags uint16, dst net.IP) net.IP {
	if (flags & dhcpmsg.FlagBroadcast) != 0 {
		return net.IPv4bcast
	}
	return dst
}
