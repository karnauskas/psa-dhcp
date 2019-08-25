package msgtmpl

import (
	"math/rand"
	"net"
	"os"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
)

type tmpl struct {
	hostname string
	xid      uint32
	hwaddr   net.HardwareAddr
}

// creates a new message template for the given interface.
func create(iface *net.Interface) tmpl {
	t := tmpl{
		xid: rand.Uint32(),
	}
	if hn, err := os.Hostname(); err == nil {
		t.hostname = hn
	}

	t.hwaddr = make(net.HardwareAddr, len(iface.HardwareAddr))
	copy(t.hwaddr, iface.HardwareAddr)
	return t
}

// Discover returns a DHCPDISCOVER message.
func Discover(iface *net.Interface) (func() ([]byte, net.IP, net.IP), uint32) {
	// This message is broadcasted in an unconfigured state.
	// We do not yet know our own IP.
	t := create(iface)
	return func() ([]byte, net.IP, net.IP) {
		return t.request(dhcpmsg.MsgTypeDiscover, ipNone, ipBcast, nil, nil), nil, nil
	}, t.xid
}

// RequestSelecting returns a new selecting request.
func RequestSelecting(iface *net.Interface, requestedIP, serverIdentifier net.IP) (func() ([]byte, net.IP, net.IP), uint32) {
	// This message is broadcasted after we received an IP offer.
	// The client is still unconfigured but picked a server and has an IP it attempts to request.
	t := create(iface)
	return func() ([]byte, net.IP, net.IP) {
		return t.request(dhcpmsg.MsgTypeRequest, ipNone, ipBcast, requestedIP, serverIdentifier), nil, nil
	}, t.xid
}

// RequestRenewing returns a renewing request.
func RequestRenewing(iface *net.Interface, requestedIP, serverIdentifier net.IP) (func() ([]byte, net.IP, net.IP), uint32) {
	// This is an unicast message of a configured client.
	// We only supply a source (our) and destination (old server identifier) IP.
	// The ServerIdentifier and RequestedIP options must not be set in this state.
	t := create(iface)
	return func() ([]byte, net.IP, net.IP) {
		return t.request(dhcpmsg.MsgTypeRequest, requestedIP, serverIdentifier, nil, nil), requestedIP, serverIdentifier
	}, t.xid
}

// RequestRebinding returns a rebinding request.
func RequestRebinding(iface *net.Interface, requestedIP net.IP) (func() ([]byte, net.IP, net.IP), uint32) {
	// This is a broadcast message of a configured client.
	// This message is similar to the RequestRenewing message bug sent as
	// a broadcast message to all (DHCP)Servers on the network.
	t := create(iface)
	return func() ([]byte, net.IP, net.IP) {
		return t.request(dhcpmsg.MsgTypeRequest, requestedIP, ipBcast, nil, nil), nil, nil
	}, t.xid
}
