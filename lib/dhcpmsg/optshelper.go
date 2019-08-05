package dhcpmsg

import (
	"encoding/binary"
	"net"
	"time"
)

type DecodedOptions struct {
	SubnetMask         *net.IP
	Routers            *[]net.IP
	DNS                *[]net.IP
	DomainName         *string
	BroadcastAddress   *net.IP
	IPAddressLeaseTime *time.Duration
	MessageType        *uint8
	ServerIdentifier   *net.IP
	Message            *string
	RenewalTime        *time.Duration
	RebindTime         *time.Duration
	ClientIdentifier   *string
}

func DecodeOptions(opts []DHCPOpt) DecodedOptions {
	d := DecodedOptions{}
	for _, o := range opts {
		switch o.Option {
		case OptSubnetMask:
			d.SubnetMask = toV4(o.Data)
		case OptRouter:
			d.Routers = toV4A(o.Data)
		case OptDNS:
			d.DNS = toV4A(o.Data)
		case OptDomainName:
			d.DomainName = toString(o.Data)
		case OptBroadcastAddress:
			d.BroadcastAddress = toV4(o.Data)
		case OptIPAddressLeaseTime:
			d.IPAddressLeaseTime = toDuration(o.Data)
		case OptMessageType:
			d.MessageType = toUint8(o.Data)
		case OptServerIdentifier:
			d.ServerIdentifier = toV4(o.Data)
		case OptMessage:
			d.Message = toString(o.Data)
		case OptRenewalTime:
			d.RenewalTime = toDuration(o.Data)
		case OptRebindTime:
			d.RebindTime = toDuration(o.Data)
		case OptClientIdentifier:
			d.ClientIdentifier = toString(o.Data)
		}
	}
	return d
}

func toUint8(x []byte) *uint8 {
	if len(x) != 1 {
		return nil
	}
	v := uint8(x[0])
	return &v
}

func toDuration(x []byte) *time.Duration {
	if len(x) != 4 {
		return nil
	}
	d := time.Second * time.Duration(binary.BigEndian.Uint32(x))
	return &d
}

func toString(x []byte) *string {
	s := string(x)
	return &s
}

// toV4 returns a net.IP array with at most one element.
func toV4(x []byte) *net.IP {
	if v := toV4A(x); v != nil && len(*v) == 1 {
		return &(*v)[0]
	}
	return nil
}

// toV4 returns a net.IP array.
func toV4A(x []byte) *[]net.IP {
	var v []net.IP
	if len(x) >= 4 && len(x)%4 == 0 {
		for i := 0; i < len(x); i += 4 {
			v = append(v, net.IPv4(x[i+0], x[i+1], x[i+2], x[i+3]))
		}
	}
	return &v
}
