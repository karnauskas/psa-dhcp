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

	binary.BigEndian.PutUint16(b[10:], ipv4csum(b[0:ihl], 0))
	if b[9] == 0x11 && len(b) >= ihl+8 {
		binary.BigEndian.PutUint16(b[ihl+6:], udp4csum(ihl, b))
	} else {
		return fmt.Errorf("unable to calculate checksum")
	}
	return nil
}

// Calculates an IPv4 header checksum.
// This function assumes that the checksum bits are zeroed out.
func ipv4csum(b []byte, acc uint32) uint16 {
	length := len(b) - 1
	for i := 0; i < length; i += 2 {
		acc += uint32(b[i]) << 8
		acc += uint32(b[i+1])
	}
	if len(b)%2 == 1 {
		acc += uint32(b[length]) << 8
	}
	for acc > 0xFFFF {
		acc = (acc >> 16) + uint32(uint16(acc))
	}
	return ^uint16(acc)
}

// Calculates an UDP header checksum.
// The input is the full packet payload, including IP header.
func udp4csum(hlen int, b []byte) uint16 {
	length := uint32(len(b) - hlen)
	csum := pseudohdrcsum(b)
	csum += length & 0xFFFF
	csum += length >> 16
	return ipv4csum(b[hlen:], csum)
}

func pseudohdrcsum(b []byte) uint32 {
	var csum uint32
	// proto
	csum += uint32(b[9])
	// src
	csum += (uint32(b[12]) + uint32(b[14])) << 8
	csum += (uint32(b[13]) + uint32(b[15])) << 0
	// dst
	csum += (uint32(b[16]) + uint32(b[18])) << 8
	csum += (uint32(b[17]) + uint32(b[19])) << 0
	return csum
}
