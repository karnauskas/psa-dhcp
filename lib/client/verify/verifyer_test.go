package verify

import (
	"net"
	"testing"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
)

func TestVerifyGenAck(t *testing.T) {
	input := []struct {
		name          string
		xid           uint32
		msg           dhcpmsg.Message
		opts          dhcpmsg.DecodedOptions
		lastmsg       dhcpmsg.Message
		lastopts      dhcpmsg.DecodedOptions
		wantSelecting State
		wantRenewing  State
		wantRebinding State
	}{
		{
			name:          "empty struct",
			wantSelecting: Failed,
			wantRenewing:  Failed,
			wantRebinding: Failed,
		},
		{
			name: "okay",
			xid:  33,
			lastmsg: dhcpmsg.Message{
				YourIP: net.IPv4(192, 168, 1, 1),
			},
			lastopts: dhcpmsg.DecodedOptions{
				ServerIdentifier: net.IPv4(192, 168, 9, 99),
			},
			msg: dhcpmsg.Message{
				Xid:    33,
				YourIP: net.IPv4(192, 168, 1, 1),
			},
			opts: dhcpmsg.DecodedOptions{
				ServerIdentifier:       net.IPv4(192, 168, 9, 99),
				Routers:                []net.IP{net.IPv4(192, 168, 0, 1)},
				MessageType:            dhcpmsg.MsgTypeAck,
				IPAddressLeaseDuration: 1 * time.Minute,
			},
			wantSelecting: Passed,
			wantRenewing:  Passed,
			wantRebinding: Passed,
		},
		{
			name: "bad xid",
			xid:  66,
			lastmsg: dhcpmsg.Message{
				YourIP: net.IPv4(192, 168, 1, 1),
			},
			lastopts: dhcpmsg.DecodedOptions{
				ServerIdentifier: net.IPv4(192, 168, 9, 99),
			},
			msg: dhcpmsg.Message{
				Xid:    33,
				YourIP: net.IPv4(192, 168, 1, 1),
			},
			opts: dhcpmsg.DecodedOptions{
				ServerIdentifier:       net.IPv4(192, 168, 9, 99),
				Routers:                []net.IP{net.IPv4(192, 168, 0, 1)},
				MessageType:            dhcpmsg.MsgTypeAck,
				IPAddressLeaseDuration: 1 * time.Minute,
			},
			wantSelecting: Failed,
			wantRenewing:  Failed,
			wantRebinding: Failed,
		},
		{
			name: "diverged server identifier",
			xid:  33,
			lastmsg: dhcpmsg.Message{
				YourIP: net.IPv4(192, 168, 1, 1),
			},
			lastopts: dhcpmsg.DecodedOptions{},
			msg: dhcpmsg.Message{
				Xid:    33,
				YourIP: net.IPv4(192, 168, 1, 1),
			},
			opts: dhcpmsg.DecodedOptions{
				ServerIdentifier:       net.IPv4(192, 168, 9, 99),
				Routers:                []net.IP{net.IPv4(192, 168, 0, 1)},
				MessageType:            dhcpmsg.MsgTypeAck,
				IPAddressLeaseDuration: 1 * time.Minute,
			},
			wantSelecting: Failed,
			wantRenewing:  Failed,
			wantRebinding: Failed,
		},
		{
			name: "diverged IP",
			xid:  33,
			lastmsg: dhcpmsg.Message{
				YourIP: net.IPv4(192, 168, 9, 1),
			},
			lastopts: dhcpmsg.DecodedOptions{
				ServerIdentifier: net.IPv4(192, 168, 9, 99),
			},
			msg: dhcpmsg.Message{
				Xid:    33,
				YourIP: net.IPv4(192, 168, 1, 1),
			},
			opts: dhcpmsg.DecodedOptions{
				ServerIdentifier:       net.IPv4(192, 168, 9, 99),
				Routers:                []net.IP{net.IPv4(192, 168, 0, 1)},
				MessageType:            dhcpmsg.MsgTypeAck,
				IPAddressLeaseDuration: 1 * time.Minute,
			},
			wantSelecting: Failed,
			wantRenewing:  Failed,
			wantRebinding: Failed,
		},
		{
			name: "bad msg type",
			xid:  33,
			lastmsg: dhcpmsg.Message{
				YourIP: net.IPv4(192, 168, 1, 1),
			},
			lastopts: dhcpmsg.DecodedOptions{
				ServerIdentifier: net.IPv4(192, 168, 9, 99),
			},
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
			wantSelecting: Failed,
			wantRenewing:  Failed,
			wantRebinding: Failed,
		},
		{
			name:    "nack",
			xid:     33,
			lastmsg: dhcpmsg.Message{},
			lastopts: dhcpmsg.DecodedOptions{
				ServerIdentifier: net.IPv4(192, 168, 9, 99),
			},
			msg: dhcpmsg.Message{},
			opts: dhcpmsg.DecodedOptions{
				ServerIdentifier: net.IPv4(192, 168, 9, 99),
				MessageType:      dhcpmsg.MsgTypeNack,
			},
			wantSelecting: IsNack,
			wantRenewing:  IsNack,
			wantRebinding: IsNack,
		},
		{
			name:    "alien nack",
			xid:     33,
			lastmsg: dhcpmsg.Message{},
			lastopts: dhcpmsg.DecodedOptions{
				ServerIdentifier: net.IPv4(192, 168, 9, 99),
			},
			msg: dhcpmsg.Message{},
			opts: dhcpmsg.DecodedOptions{
				ServerIdentifier: net.IPv4(192, 168, 100, 99),
				MessageType:      dhcpmsg.MsgTypeNack,
			},
			wantSelecting: Failed,
			wantRenewing:  Failed,
			wantRebinding: Failed,
		},
	}

	for _, test := range input {
		if got := VerifySelectingAck(test.lastmsg, test.lastopts, test.xid)(test.msg, test.opts); got != test.wantSelecting {
			t.Errorf("VerifySelectingAck(%s) = %d, want %d", test.name, got, test.wantSelecting)
		}
		if got := VerifyRenewingAck(test.lastmsg, test.lastopts, test.xid)(test.msg, test.opts); got != test.wantRenewing {
			t.Errorf("VerifyRenewingAck(%s) = %d, want %d", test.name, got, test.wantRenewing)
		}
		if got := VerifyRebindingAck(test.lastmsg, test.lastopts, test.xid)(test.msg, test.opts); got != test.wantRebinding {
			t.Errorf("VerifyRebindingAck(%s) = %d, want %d", test.name, got, test.wantRebinding)
		}
	}
}
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
		if got := VerifyOffer(test.xid)(test.msg, test.opts); got != test.want {
			t.Errorf("VerifyOffer(%s) = %d, want %d", test.name, got, test.want)
		}
	}
}
