package verify

import (
	"net"
	"testing"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
)

func TestVerifyOffer(t *testing.T) {
	input := []struct {
		name string
		xid  uint32
		msg  dhcpmsg.Message
		opts dhcpmsg.DecodedOptions
		want State
	}{
		{
			name: "empty struct",
			want: Failed,
		},
		{
			name: "okay",
			xid:  33,
			msg: dhcpmsg.Message{
				Xid:    33,
				YourIP: net.IPv4(192, 168, 1, 1),
			},
			opts: dhcpmsg.DecodedOptions{
				ServerIdentifier:       net.IPv4(192, 168, 9, 99),
				Routers:                []net.IP{net.IPv4(192, 168, 0, 1)},
				MessageType:            dhcpmsg.MsgTypeOffer,
				IPAddressLeaseDuration: 1 * time.Minute,
			},
			want: Passed,
		},
		{
			name: "no routers",
			xid:  33,
			msg: dhcpmsg.Message{
				Xid:    33,
				YourIP: net.IPv4(192, 168, 1, 1),
			},
			opts: dhcpmsg.DecodedOptions{
				ServerIdentifier:       net.IPv4(192, 168, 9, 99),
				MessageType:            dhcpmsg.MsgTypeOffer,
				IPAddressLeaseDuration: 1 * time.Minute,
			},
			want: Failed,
		},
		{
			name: "no ip",
			xid:  33,
			msg: dhcpmsg.Message{
				Xid: 33,
			},
			opts: dhcpmsg.DecodedOptions{
				ServerIdentifier:       net.IPv4(192, 168, 9, 99),
				Routers:                []net.IP{net.IPv4(192, 168, 0, 1)},
				MessageType:            dhcpmsg.MsgTypeOffer,
				IPAddressLeaseDuration: 1 * time.Minute,
			},
			want: Failed,
		},
		{
			name: "bad type",
			xid:  33,
			msg: dhcpmsg.Message{
				Xid:    33,
				YourIP: net.IPv4(192, 168, 1, 1),
			},
			opts: dhcpmsg.DecodedOptions{
				ServerIdentifier:       net.IPv4(192, 168, 9, 99),
				Routers:                []net.IP{net.IPv4(192, 168, 0, 1)},
				MessageType:            dhcpmsg.MsgTypeRequest,
				IPAddressLeaseDuration: 1 * time.Minute,
			},
			want: Failed,
		},
		{
			name: "xid mismatch",
			xid:  66,
			msg: dhcpmsg.Message{
				Xid:    33,
				YourIP: net.IPv4(192, 168, 1, 1),
			},
			opts: dhcpmsg.DecodedOptions{
				ServerIdentifier:       net.IPv4(192, 168, 9, 99),
				Routers:                []net.IP{net.IPv4(192, 168, 0, 1)},
				MessageType:            dhcpmsg.MsgTypeOffer,
				IPAddressLeaseDuration: 1 * time.Minute,
			},
			want: Failed,
		},
		{
			name: "no server identifier",
			xid:  33,
			msg: dhcpmsg.Message{
				Xid:    33,
				YourIP: net.IPv4(192, 168, 1, 1),
			},
			opts: dhcpmsg.DecodedOptions{
				Routers:                []net.IP{net.IPv4(192, 168, 0, 1)},
				MessageType:            dhcpmsg.MsgTypeOffer,
				IPAddressLeaseDuration: 1 * time.Minute,
			},
			want: Failed,
		},
		{
			name: "bad server identifier",
			xid:  33,
			msg: dhcpmsg.Message{
				Xid:    33,
				YourIP: net.IPv4(192, 168, 1, 1),
			},
			opts: dhcpmsg.DecodedOptions{
				ServerIdentifier:       net.IPv4(0, 0, 0, 0),
				Routers:                []net.IP{net.IPv4(192, 168, 0, 1)},
				MessageType:            dhcpmsg.MsgTypeOffer,
				IPAddressLeaseDuration: 1 * time.Minute,
			},
			want: Failed,
		},
	}

	for _, test := range input {
		got := VerifyOffer(test.xid)(test.msg, test.opts)
		if got != test.want {
			t.Errorf("VerifyOffer(%s) = %d, want %d", test.name, got, test.want)
		}
	}
}
