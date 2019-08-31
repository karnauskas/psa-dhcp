package replies

import (
	"net"

	"gitlab.com/adrian_blx/psa-dhcp/lib/layer"
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
