package dhcpmsg

import (
	"encoding/binary"
	"net"
	"time"
)

type DecodedOptions struct {
	MessageType            uint8
	MaxMessageSize         uint16
	InterfaceMTU           uint16
	RequestedIP            net.IP
	ServerIdentifier       net.IP
	BroadcastAddress       net.IP
	SubnetMask             net.IPMask
	Routers                []net.IP
	DNS                    []net.IP
	IPAddressLeaseDuration time.Duration
	RenewalDuration        time.Duration
	RebindDuration         time.Duration
	DomainName             string
	ClientIdentifier       string
	Message                string
	ParametersList         []uint8
}

func DecodeOptions(opts []DHCPOpt) DecodedOptions {
	d := DecodedOptions{}
	for _, o := range opts {
		switch o.Option {
		case OptSubnetMask:
			d.SubnetMask = toNetmask(o.Data)
		case OptRouter:
			d.Routers = toV4A(o.Data)
		case OptDNS:
			d.DNS = toV4A(o.Data)
		case OptDomainName:
			d.DomainName = toString(o.Data)
		case OptBroadcastAddress:
			d.BroadcastAddress = toV4(o.Data)
		case OptRequestedIP:
			d.RequestedIP = toV4(o.Data)
		case OptIPAddressLeaseDuration:
			d.IPAddressLeaseDuration = toDuration(o.Data)
		case OptMessageType:
			d.MessageType = toUint8(o.Data)
		case OptMaxMessageSize:
			d.MaxMessageSize = toUint16(o.Data)
		case OptInterfaceMTU:
			d.InterfaceMTU = toUint16(o.Data)
		case OptServerIdentifier:
			d.ServerIdentifier = toV4(o.Data)
		case OptMessage:
			d.Message = toString(o.Data)
		case OptRenewalDuration:
			d.RenewalDuration = toDuration(o.Data)
		case OptRebindDuration:
			d.RebindDuration = toDuration(o.Data)
		case OptClientIdentifier:
			d.ClientIdentifier = toString(o.Data)
		case OptParametersList:
			d.ParametersList = o.Data

		}
	}
	return d
}

func toUint8(x []byte) (v uint8) {
	if len(x) == 1 {
		v = uint8(x[0])
	}
	return
}

func toUint16(x []byte) (v uint16) {
	if len(x) == 2 {
		v = uint16(x[0])<<8 | uint16(x[1])
	}
	return
}

func toDuration(x []byte) (d time.Duration) {
	if len(x) == 4 {
		d = time.Second * time.Duration(binary.BigEndian.Uint32(x))
	}
	return
}

func toString(x []byte) string {
	return string(x)
}

func toNetmask(x []byte) net.IPMask {
	if len(x) != 4 {
		return nil
	}
	return net.IPv4Mask(x[0], x[1], x[2], x[3])
}

// toV4 returns a net.IP array with at most one element.
func toV4(x []byte) (ip net.IP) {
	if v := toV4A(x); len(v) == 1 {
		ip = v[0]
	}
	return
}

// toV4 returns a net.IP array.
func toV4A(x []byte) []net.IP {
	var v []net.IP
	if len(x) >= 4 && len(x)%4 == 0 {
		for i := 0; i < len(x); i += 4 {
			v = append(v, net.IPv4(x[i+0], x[i+1], x[i+2], x[i+3]))
		}
	}
	return v
}
