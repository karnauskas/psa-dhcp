package layer

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	ARPOpRequest = 1
)

type ARP struct {
	SenderMAC net.HardwareAddr
	SenderIP  net.IP
	TargetMAC net.HardwareAddr
	TargetIP  net.IP
	Opcode    uint8
}

func (u ARP) Assemble() []byte {
	b := make([]byte, 28)
	binary.BigEndian.PutUint32(b[0:], 0x00010800) // ETH + IPv4
	binary.BigEndian.PutUint32(b[4:], 0x06040000) // HWsz, Psz, Opcode
	b[7] = u.Opcode

	copy(b[8:], u.SenderMAC[:])
	if v4 := u.SenderIP.To4(); v4 != nil {
		copy(b[14:], v4)
	}
	copy(b[18:], u.TargetMAC[:])
	if v4 := u.TargetIP.To4(); v4 != nil {
		copy(b[24:], v4)
	}
	return b
}

func DecodeARP(b []byte) (*ARP, error) {
	if len(b) != 28 {
		return nil, fmt.Errorf("short arp")
	}

	arp := &ARP{
		Opcode:    b[7],
		SenderMAC: b[8:14],
		SenderIP:  net.IPv4(b[14], b[15], b[16], b[17]),
		TargetMAC: b[18:24],
		TargetIP:  net.IPv4(b[24], b[25], b[26], b[27]),
	}
	return arp, nil
}

/*
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
*/
