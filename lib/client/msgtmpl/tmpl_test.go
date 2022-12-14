package msgtmpl

import (
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"git.sr.ht/~adrian-blx/psa-dhcp/lib/dhcpmsg"
	"git.sr.ht/~adrian-blx/psa-dhcp/lib/layer"
)

var (
	testSource     = net.IPv4(1, 2, 3, 9)
	testIdentifier = net.IPv4(21, 23, 24, 25)
	testIface      = net.Interface{HardwareAddr: []byte{1, 2, 3, 4, 5, 6}}
	testMAC        = net.HardwareAddr{1, 2, 3, 4, 5, 6}
	testParams     = dhcpmsg.OptionParametersList(
		dhcpmsg.OptSubnetMask, dhcpmsg.OptRouter, dhcpmsg.OptIPAddressLeaseDuration,
		dhcpmsg.OptServerIdentifier, dhcpmsg.OptDNS, dhcpmsg.OptDomainName,
		dhcpmsg.OptInterfaceMTU, dhcpmsg.OptRenewalDuration, dhcpmsg.OptRebindDuration)
)

type bundle struct {
	IPv4 layer.IPv4
	UDP  layer.UDP
	Msg  dhcpmsg.Message
}

func TestDiscover(t *testing.T) {
	want := bundle{
		IPv4: layer.IPv4{
			TTL:      64,
			Protocol: 0x11,
			// we are not configured yet.
			Source: ipNone,
			// don't know any servers.
			Destination: ipBcast,
		},
		UDP: layer.UDP{
			SrcPort: 68,
			DstPort: 67,
		},
		Msg: dhcpmsg.Message{
			Op:        dhcpmsg.OpRequest,
			Htype:     dhcpmsg.HtypeETHER,
			ClientIP:  ipNone,
			YourIP:    ipNone,
			NextIP:    ipNone,
			RelayIP:   ipNone,
			ClientMAC: testMAC,
			Cookie:    dhcpmsg.DHCPCookie,
			Options: []dhcpmsg.DHCPOpt{
				dhcpmsg.OptionType(dhcpmsg.MsgTypeDiscover),
				dhcpmsg.OptionClientIdentifier(testMAC),
				dhcpmsg.OptionMaxMessageSize(maxMsgSize),
				testParams,
				dhcpmsg.OptionHostname("TEST"),
			},
		},
	}

	rq, xid := Discover(&testIface)
	want.Msg.Xid = xid
	data, _, _ := rq()

	got, err := undo(data)
	if err != nil {
		t.Errorf("TestRequestSelecting = %v, want nil err", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TestRequestSelecting had diff: %s", diff)
	}
}

func TestRequestSelecing(t *testing.T) {
	want := bundle{
		IPv4: layer.IPv4{
			TTL:      64,
			Protocol: 0x11,
			// we are not configured yet.
			Source: ipNone,
			// selected a server, but must broadcast.
			Destination: ipBcast,
		},
		UDP: layer.UDP{
			SrcPort: 68,
			DstPort: 67,
		},
		Msg: dhcpmsg.Message{
			Op:        dhcpmsg.OpRequest,
			Htype:     dhcpmsg.HtypeETHER,
			ClientIP:  ipNone,
			YourIP:    ipNone,
			NextIP:    ipNone,
			RelayIP:   ipNone,
			ClientMAC: testMAC,
			Cookie:    dhcpmsg.DHCPCookie,
			Options: []dhcpmsg.DHCPOpt{
				dhcpmsg.OptionType(dhcpmsg.MsgTypeRequest),
				dhcpmsg.OptionClientIdentifier(testMAC),
				dhcpmsg.OptionMaxMessageSize(maxMsgSize),
				testParams,
				dhcpmsg.OptionRequestedIP(testSource),
				dhcpmsg.OptionServerIdentifier(testIdentifier),
				dhcpmsg.OptionHostname("TEST"),
			},
		},
	}

	rq, xid := RequestSelecting(&testIface, testSource, testIdentifier)
	want.Msg.Xid = xid
	data, _, _ := rq()

	got, err := undo(data)
	if err != nil {
		t.Errorf("TestRequestSelecting = %v, want nil err", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TestRequestSelecting had diff: %s", diff)
	}
}

func TestRequestRenewing(t *testing.T) {
	want := bundle{
		IPv4: layer.IPv4{
			TTL:      64,
			Protocol: 0x11,
			Source:   testSource,
			// This is NOT a broadcast message.
			Destination: testIdentifier,
		},
		UDP: layer.UDP{
			SrcPort: 68,
			DstPort: 67,
		},
		Msg: dhcpmsg.Message{
			Op:        dhcpmsg.OpRequest,
			Htype:     dhcpmsg.HtypeETHER,
			ClientIP:  testSource,
			YourIP:    ipNone,
			NextIP:    ipNone,
			RelayIP:   ipNone,
			ClientMAC: testMAC,
			Cookie:    dhcpmsg.DHCPCookie,
			Options: []dhcpmsg.DHCPOpt{
				dhcpmsg.OptionType(dhcpmsg.MsgTypeRequest),
				dhcpmsg.OptionClientIdentifier(testMAC),
				dhcpmsg.OptionMaxMessageSize(maxMsgSize),
				testParams,
				dhcpmsg.OptionHostname("TEST"),
			},
		},
	}

	rq, xid := RequestRenewing(&testIface, testSource, testIdentifier)
	want.Msg.Xid = xid
	data, _, _ := rq()

	got, err := undo(data)
	if err != nil {
		t.Errorf("TestRequestRenewing = %v, want nil err", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TestRequestRenewing had diff: %s", diff)
	}
}

func TestRequestRebinding(t *testing.T) {
	// Same as renewing, but sent to the broadcast address.
	want := bundle{
		IPv4: layer.IPv4{
			TTL:         64,
			Protocol:    0x11,
			Source:      testSource,
			Destination: ipBcast,
		},
		UDP: layer.UDP{
			SrcPort: 68,
			DstPort: 67,
		},
		Msg: dhcpmsg.Message{
			Op:        dhcpmsg.OpRequest,
			Htype:     dhcpmsg.HtypeETHER,
			ClientIP:  testSource,
			YourIP:    ipNone,
			NextIP:    ipNone,
			RelayIP:   ipNone,
			ClientMAC: testMAC,
			Cookie:    dhcpmsg.DHCPCookie,
			Options: []dhcpmsg.DHCPOpt{
				dhcpmsg.OptionType(dhcpmsg.MsgTypeRequest),
				dhcpmsg.OptionClientIdentifier(testMAC),
				dhcpmsg.OptionMaxMessageSize(maxMsgSize),
				testParams,
				dhcpmsg.OptionHostname("TEST"),
			},
		},
	}

	rq, xid := RequestRebinding(&testIface, testSource)
	want.Msg.Xid = xid
	data, _, _ := rq()

	got, err := undo(data)
	if err != nil {
		t.Errorf("TestRequestBinding = %v, want nil err", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("TestRequestBinding had diff: %s", diff)
	}
}

func undo(raw []byte) (bundle, error) {
	v4, err := layer.DecodeIPv4(raw)
	if err != nil {
		return bundle{}, err
	}

	udp, err := layer.DecodeUDP(v4.Data)
	if err != nil {
		return bundle{}, err
	}

	msg, err := dhcpmsg.Decode(udp.Data)
	if err != nil {
		return bundle{}, err
	}

	v4.Identification = 0
	v4.Checksum = 0
	v4.Data = nil
	udp.Data = nil

	// Remove hostname to have this stable.
	var opts []dhcpmsg.DHCPOpt
	for _, o := range msg.Options {
		if o.Option != dhcpmsg.OptHostname {
			opts = append(opts, o)
		}
	}
	opts = append(opts, dhcpmsg.OptionHostname("TEST"))
	msg.Options = opts

	return bundle{IPv4: *v4, UDP: *udp, Msg: *msg}, nil
}
