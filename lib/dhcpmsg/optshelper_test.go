package dhcpmsg

import (
	"net"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestDecodeOptions(t *testing.T) {
	input := []struct {
		name string
		data []DHCPOpt
		want DecodedOptions
	}{
		{
			name: "empty",
			data: []DHCPOpt{},
			want: DecodedOptions{},
		}, {
			name: "noopts",
			data: []DHCPOpt{
				{Option: OptPadding, Data: []byte{}},
				{Option: OptEnd, Data: []byte{}},
			},
			want: DecodedOptions{},
		}, {
			name: "IPs",
			data: []DHCPOpt{
				{Option: OptSubnetMask, Data: []byte{0xff, 0xff, 0xfe, 0x00}},
				{Option: OptRouter, Data: []byte{192, 168, 1, 1, 192, 168, 1, 2}},
				{Option: OptDNS, Data: []byte{192, 168, 2, 1, 192, 168, 4, 9}},
				{Option: OptBroadcastAddress, Data: []byte{0xff, 0xff, 0xfd, 0xfb}},
				{Option: OptRequestedIP, Data: []byte{192, 168, 1, 99}},
				{Option: OptServerIdentifier, Data: []byte{192, 168, 100, 200}},
			},
			want: DecodedOptions{
				SubnetMask:       net.IPv4Mask(255, 255, 254, 0),
				Routers:          []net.IP{net.IPv4(192, 168, 1, 1), net.IPv4(192, 168, 1, 2)},
				DNS:              []net.IP{net.IPv4(192, 168, 2, 1), net.IPv4(192, 168, 4, 9)},
				BroadcastAddress: net.IPv4(255, 255, 253, 251),
				RequestedIP:      net.IPv4(192, 168, 1, 99),
				ServerIdentifier: net.IPv4(192, 168, 100, 200),
			},
		}, {
			name: "ints",
			data: []DHCPOpt{
				{Option: OptMessageType, Data: []byte{0xfe}},
			},
			want: DecodedOptions{
				MessageType: 254,
			},
		}, {
			name: "strings",
			data: []DHCPOpt{
				{Option: OptDomainName, Data: []byte{'f', 'o', 'o'}},
				{Option: OptMessage, Data: []byte{'x', 'x', 'y', 'y', 'z', 'z'}},
				{Option: OptClientIdentifier, Data: []byte{'a', 'b', 'c', 'd'}},
			},
			want: DecodedOptions{
				DomainName:       "foo",
				Message:          "xxyyzz",
				ClientIdentifier: "abcd",
			},
		}, {
			name: "time",
			data: []DHCPOpt{
				{Option: OptIPAddressLeaseTime, Data: []byte{0x01, 0x02, 0x03, 0x10}},
				{Option: OptRenewalTime, Data: []byte{0x11, 0x02, 0x03, 0x10}},
				{Option: OptRebindTime, Data: []byte{0x21, 0x02, 0x03, 0x10}},
			},
			want: DecodedOptions{
				IPAddressLeaseTime: time.Second * 0x01020310,
				RenewalTime:        time.Second * 0x11020310,
				RebindTime:         time.Second * 0x21020310,
			},
		},
	}

	for _, test := range input {
		decoded := DecodeOptions(test.data)
		if diff := cmp.Diff(test.want, decoded); diff != "" {
			t.Errorf("test '%s' had a diff: %s", test.name, diff)
		}
	}
}
