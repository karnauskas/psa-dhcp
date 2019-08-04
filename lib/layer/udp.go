package layer

import (
	"encoding/binary"
	"fmt"
)

const (
	udpHlen = 8
)

type UDP struct {
	SrcPort uint16
	DstPort uint16
	Data    []byte
}

func (u UDP) Assemble() []byte {
	dlen := len(u.Data)
	b := make([]byte, udpHlen+dlen)
	binary.BigEndian.PutUint16(b[0:], u.SrcPort)
	binary.BigEndian.PutUint16(b[2:], u.DstPort)
	binary.BigEndian.PutUint16(b[4:], uint16(udpHlen+dlen))
	// 6:7 = checksum
	copy(b[8:], u.Data)
	return b
}

func DecodeUDP(b []byte) (*UDP, error) {
	plen := len(b)
	if plen < udpHlen {
		return nil, fmt.Errorf("short udp")
	}
	tlen := int(binary.BigEndian.Uint16(b[4:]))
	if tlen != plen {
		return nil, fmt.Errorf("truncated udp")
	}

	udp := &UDP{
		SrcPort: binary.BigEndian.Uint16(b[0:]),
		DstPort: binary.BigEndian.Uint16(b[2:]),
		Data:    b[udpHlen:],
	}
	return udp, nil
}
