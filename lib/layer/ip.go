package layer

import (
	"encoding/binary"
	"net"
)

type IPv4 struct {
	Identification uint16
	Flags          uint16
	TTL            uint8
	Protocol       uint8
	Checksum       uint16
	Source         net.IP
	Destination    net.IP
	Data           []byte
}

func (h IPv4) Assemble() []byte {
	hlen := 20
	dlen := len(h.Data)
	b := make([]byte, hlen+dlen)

	b[0] = byte(4<<4) + byte(hlen>>2)                    // Version + IHL
	b[1] = 0x00                                          // DSF
	binary.BigEndian.PutUint16(b[2:], uint16(hlen+dlen)) // Total length
	binary.BigEndian.PutUint16(b[4:], h.Identification)
	binary.BigEndian.PutUint16(b[6:], h.Flags)
	b[8] = h.TTL
	b[9] = h.Protocol
	// checksum 10, 11

	if h.Source != nil {
		if x := h.Source.To4(); x != nil {
			copy(b[12:], x[0:4])
		}
	}
	if h.Destination != nil {
		if x := h.Destination.To4(); x != nil {
			copy(b[16:], x[0:4])
		}
	}
	copy(b[20:], h.Data)

	setV4Checksum(b)
	return b
}
