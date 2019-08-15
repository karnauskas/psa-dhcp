package msgtmpl

import (
	"net"
	"os"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
)

type tmpl struct {
	hostname string
	start    time.Time
	xid      uint32
	hwaddr   [6]byte
}

// New returns a new message template for the given interface.
func New(iface *net.Interface, xid uint32) tmpl {
	t := tmpl{
		start: time.Now(),
		xid:   xid,
	}
	if hn, err := os.Hostname(); err == nil {
		t.hostname = hn
	}
	if len(iface.HardwareAddr) == 6 {
		copy(t.hwaddr[:], iface.HardwareAddr[:])
	}
	return t
}

// Discover returns a DHCPDISCOVER message.
func (rx *tmpl) Discover() []byte {
	// This message is broadcasted in an unconfigured state.
	// We do not yet know our own IP.
	return rx.request(dhcpmsg.MsgTypeDiscover, ipNone, ipBcast, nil, nil)
}

// RequestSelecting returns a new selecting request.
func (rx *tmpl) RequestSelecting(requestedIP, serverIdentifier net.IP) []byte {
	// This message is broadcasted after we received an IP offer.
	// The client is still unconfigured but picked a server and has an IP it attempts to request.
	return rx.request(dhcpmsg.MsgTypeRequest, ipNone, ipBcast, &requestedIP, &serverIdentifier)
}

// RequestRenewing returns a renewing request.
func (rx *tmpl) RequestRenewing(requestedIP, serverIdentifier net.IP) []byte {
	// This is an unicast message of a configured client.
	// We only supply a source (our) and destination (old server identifier) IP.
	// The ServerIdentifier and RequestedIP options must not be set in this state.
	return rx.request(dhcpmsg.MsgTypeRequest, requestedIP, serverIdentifier, nil, nil)
}

// RequestRebinding returns a rebinding request.
func (rx *tmpl) RequestRebinding(requestedIP net.IP) []byte {
	// This is a broadcast message of a configured client.
	// This message is similar to the RequestRenewing message bug sent as
	// a broadcast message to all (DHCP)Servers on the network.
	return rx.request(dhcpmsg.MsgTypeRequest, requestedIP, ipBcast, nil, nil)
}
