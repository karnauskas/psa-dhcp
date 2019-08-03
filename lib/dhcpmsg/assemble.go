package dhcpmsg

import (
	"net"
)

type Message struct {
	Op             uint8
	Htype          uint8
	Hlen           uint8
	Hops           uint8
	Xid            uint32
	Secs           uint16
	Flags          uint16
	ClientIP       net.IP
	YourIP         net.IP
	NextIP         net.IP
	RelayIP        net.IP
	ClientMAC      [6]byte
	MACPadding     [10]byte
	ServerHostName [64]byte
	BootFilename   [128]byte
	Cookie         uint32
	Options        []DHCPOpt
}

func (msg Message) Assemble() []byte {
	buf := make([]byte, 240)
	buf[0] = msg.Op
	buf[1] = msg.Htype
	buf[2] = msg.Hlen
	buf[3] = msg.Hops
	setU32Int(buf[4:], msg.Xid)
	setU16Int(buf[8:], msg.Secs)
	setU16Int(buf[10:], msg.Flags)
	setIPv4(buf[12:], msg.ClientIP)
	setIPv4(buf[16:], msg.YourIP)
	setIPv4(buf[20:], msg.NextIP)
	setIPv4(buf[24:], msg.RelayIP)
	copy(buf[28:], msg.ClientMAC[:])
	copy(buf[34:], msg.MACPadding[:])
	copy(buf[44:], msg.ServerHostName[:])
	copy(buf[108:], msg.BootFilename[:])
	setU32Int(buf[236:], 0x63825363) // DHCP

	for _, opt := range msg.Options {
		b := make([]byte, 2+len(opt.data))
		b[0] = opt.option
		// FIXME: Check overflow.
		b[1] = uint8(len(opt.data))
		copy(b[2:], opt.data[:])
		buf = append(buf, b...)
	}
	if len(msg.Options) > 0 {
		buf = append(buf, 0xff)
	}
	return buf
}

const (
	FlagBroadcast = 1 << 15
)

func setU16Int(b []byte, val uint16) {
	b[0] = byte(val >> 8)
	b[1] = byte(val & 0xFF)
}

func setU32Int(b []byte, val uint32) {
	b[0] = byte(val >> 24)
	b[1] = byte(val >> 16)
	b[2] = byte(val >> 8)
	b[3] = byte(val)
}

func setIPv4(b []byte, ip net.IP) {
	if v4 := ip.To4(); v4 != nil {
		copy(b, v4)
	}
}

type DHCPOpt struct {
	option uint8
	data   []byte
}

func OptDiscover() DHCPOpt {
	return DHCPOpt{option: 53, data: []byte{1}}
}

func OptHostname(n string) DHCPOpt {
	return DHCPOpt{option: 12, data: []byte(n)}
}
