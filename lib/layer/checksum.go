package layer

import (
	"encoding/binary"
	"fmt"
)

func setV4Checksum(b []byte) error {
	if len(b) < 1 || b[0]>>4 != 4 {
		return fmt.Errorf("not a V4 packet")
	}

	ihl := int((b[0] & 0xF) << 2)
	if ihl < 20 || len(b) < ihl {
		return fmt.Errorf("short packet")
	}

	binary.BigEndian.PutUint16(b[10:], ipv4csum(b[0:ihl]))
	return nil
}

// calculates an IPv4 header checksum.
// This function assumes that the checksum bits are zeroed out.
func ipv4csum(b []byte) uint16 {
	var acc uint32
	for i := 0; i < len(b); i += 2 {
		acc += uint32(b[i]) << 8
		acc += uint32(b[i+1])
	}
	for acc > 0xFFFF {
		acc = (acc >> 16) + uint32(uint16(acc))
	}
	return ^uint16(acc)
}
