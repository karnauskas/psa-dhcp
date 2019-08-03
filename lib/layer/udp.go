package layer

import (
	"encoding/binary"
)

type UDP struct {
	SrcPort uint16
	DstPort uint16
	Data    []byte
}

func (u UDP) Assemble() []byte {
	hlen := 8
	dlen := len(u.Data)
	b := make([]byte, hlen+dlen)
	binary.BigEndian.PutUint16(b[0:], u.SrcPort)
	binary.BigEndian.PutUint16(b[2:], u.DstPort)
	binary.BigEndian.PutUint16(b[4:], uint16(hlen+dlen))
	// 6:7 = checksum
	copy(b[8:], u.Data)
	return b
}
