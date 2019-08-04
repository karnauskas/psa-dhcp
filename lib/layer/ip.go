package layer

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	ipv4Hlen = 20
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
	dlen := len(h.Data)
	b := make([]byte, ipv4Hlen+dlen)

	b[0] = byte(4<<4) + byte(ipv4Hlen>>2)                    // Version + IHL
	b[1] = 0x00                                              // DSF
	binary.BigEndian.PutUint16(b[2:], uint16(ipv4Hlen+dlen)) // Total length
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

func DecodeIPv4(b []byte) (*IPv4, error) {
	plen := len(b)
	if plen < ipv4Hlen {
		return nil, fmt.Errorf("short ipv4")
	}

	version := b[0] >> 4
	ihl := int((b[0] & 0xF) << 2)
	if version != 4 || plen < ihl || ihl < ipv4Hlen {
		return nil, fmt.Errorf("invalid packet")
	}

	tlen := int(binary.BigEndian.Uint16(b[2:]))
	if tlen != plen {
		return nil, fmt.Errorf("truncated packet: %d != %d", tlen, plen)
	}

	v4 := &IPv4{}
	v4.Identification = binary.BigEndian.Uint16(b[4:])
	v4.Flags = binary.BigEndian.Uint16(b[6:])
	v4.TTL = b[8]
	v4.Protocol = b[9]
	v4.Checksum = binary.BigEndian.Uint16(b[10:])
	v4.Source = net.IPv4(b[12], b[13], b[14], b[15])
	v4.Destination = net.IPv4(b[16], b[17], b[18], b[19])
	v4.Data = b[ihl:tlen]
	return v4, nil
}
