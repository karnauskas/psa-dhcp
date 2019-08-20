package dhcpmsg

import (
	"encoding/binary"
	"hash/crc32"
	"net"
	"time"
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

type DHCPOpt struct {
	Option uint8
	Data   []byte
}

// Assemble assembles a dhcp message into raw bytes.
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
	setU32Int(buf[236:], msg.Cookie)

	for _, opt := range msg.Options {
		b := make([]byte, 2+len(opt.Data))
		b[0] = opt.Option
		// FIXME: Check overflow.
		b[1] = uint8(len(opt.Data))
		copy(b[2:], opt.Data[:])
		buf = append(buf, b...)
	}
	if len(msg.Options) > 0 {
		buf = append(buf, 0xff)
	}
	return buf
}

func setU16Int(b []byte, val uint16) {
	binary.BigEndian.PutUint16(b, val)
}

func setU32Int(b []byte, val uint32) {
	binary.BigEndian.PutUint32(b, val)
}

func setIPv4(b []byte, ip net.IP) {
	if v4 := ip.To4(); v4 != nil {
		copy(b, v4)
	}
}

func OptionType(t uint8) DHCPOpt {
	return DHCPOpt{Option: OptMessageType, Data: []byte{t}}
}

func OptionHostname(n string) DHCPOpt {
	return DHCPOpt{Option: OptHostname, Data: []byte(n)}
}

func OptionServerIdentifier(ip net.IP) DHCPOpt {
	return optIP(OptServerIdentifier, ip)
}

func OptionRequestedIP(ip net.IP) DHCPOpt {
	return optIP(OptRequestedIP, ip)
}

func OptionRouter(ip net.IP) DHCPOpt {
	return optIP(OptRouter, ip)
}

func optIP(ot uint8, ip net.IP) DHCPOpt {
	b := make([]byte, 4)
	if x := ip.To4(); x != nil {
		copy(b[0:4], x[0:4])
	}
	return DHCPOpt{Option: ot, Data: b}
}

func OptionMaxMessageSize(size uint16) DHCPOpt {
	data := make([]byte, 2)
	setU16Int(data, size)
	return DHCPOpt{Option: OptMaxMessageSize, Data: data}
}

func OptionInterfaceMTU(size uint16) DHCPOpt {
	data := make([]byte, 2)
	setU16Int(data, size)
	return DHCPOpt{Option: OptInterfaceMTU, Data: data}
}

func OptionClientIdentifier(hwaddr [6]byte) DHCPOpt {
	id := make([]byte, 15)
	setU32Int(id[1:5], crc32.ChecksumIEEE(hwaddr[0:])) // IAID
	copy(id[9:15], hwaddr[0:6])
	id[0] = 0xff // Type.
	id[6] = 3    // Link layer without time
	id[8] = 1    // Ethernet
	return DHCPOpt{Option: OptClientIdentifier, Data: id}
}

func OptionParametersList(params ...uint8) DHCPOpt {
	l := make([]byte, len(params))
	for i, p := range params {
		l[i] = p
	}
	return DHCPOpt{Option: OptParametersList, Data: l}
}

func OptionIPAddressLeaseDuration(d time.Duration) DHCPOpt {
	b := make([]byte, 4)
	setU32Int(b, uint32(d.Seconds()))
	return DHCPOpt{Option: OptIPAddressLeaseDuration, Data: b}
}

func OptionSubnetMask(mask net.IPMask) DHCPOpt {
	return DHCPOpt{Option: OptSubnetMask, Data: mask}
}
