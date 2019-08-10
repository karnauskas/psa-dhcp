package dhcpmsg

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	dhcpMinLen = 240
)

func Decode(b []byte) (*Message, error) {
	plen := len(b)
	if plen < dhcpMinLen {
		return nil, fmt.Errorf("short dhcpmsg")
	}

	msg := &Message{
		Op:       b[0],
		Htype:    b[1],
		Hlen:     b[2],
		Hops:     b[3],
		Xid:      binary.BigEndian.Uint32(b[4:]),
		Secs:     binary.BigEndian.Uint16(b[8:]),
		Flags:    binary.BigEndian.Uint16(b[10:]),
		Cookie:   binary.BigEndian.Uint32(b[236:]),
		ClientIP: net.IPv4(b[12], b[13], b[14], b[15]),
		YourIP:   net.IPv4(b[16], b[17], b[18], b[19]),
		NextIP:   net.IPv4(b[20], b[21], b[22], b[23]),
		RelayIP:  net.IPv4(b[24], b[25], b[26], b[27]),
	}
	copy(msg.ClientMAC[:], b[28:])
	copy(msg.MACPadding[:], b[34:])
	copy(msg.ServerHostName[:], b[44:])
	copy(msg.BootFilename[:], b[108:])

	c := dhcpMinLen
	for c < plen {
		opt := b[c]
		if c++; opt == 0x00 {
			continue
		}
		if !(c < plen) || opt == 0xff {
			break
		}
		olen := int(b[c])
		if c++; !(c+olen <= plen) {
			break
		}
		msg.Options = append(msg.Options, DHCPOpt{Option: opt, Data: b[c : c+olen]})
		c += olen
	}
	if c != plen {
		return nil, fmt.Errorf("truncated options")
	}
	return msg, nil
}
