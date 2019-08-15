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
	return rx.request(dhcpmsg.MsgTypeDiscover, ipNone, ipBcast, nil, nil)
}

// RequestSelecting returns a new selecting request.
func (rx *tmpl) RequestSelecting(requestedIP, serverIdentifier net.IP) []byte {
	return rx.request(dhcpmsg.MsgTypeRequest, net.IPv4(0, 0, 0, 0), net.IPv4(255, 255, 255, 255),
		&requestedIP, &serverIdentifier)
}

// RequestRenewing returns a renewing request.
func (rx *tmpl) RequestRenewing(requestedIP, serverIdentifier net.IP) []byte {
	return rx.request(dhcpmsg.MsgTypeRequest, requestedIP, serverIdentifier, nil, nil)
}

// RequestRebinding returns a rebinding request.
func (rx *tmpl) RequestRebinding(requestedIP, serverIdentifier net.IP) []byte {
	return rx.request(dhcpmsg.MsgTypeRequest, requestedIP, serverIdentifier, nil, nil)
}
