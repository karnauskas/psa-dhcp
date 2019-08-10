package client

import (
	"fmt"
	"net"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
)

func verifyAck(lm dhcpmsg.Message, xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) bool {
	return func(m dhcpmsg.Message, opt dhcpmsg.DecodedOptions) bool {
		if m.Xid != xid {
			return false
		}
		if x := opt.MessageType; x == nil || *x != dhcpmsg.MsgTypeAck {
			fmt.Printf("YFAIL :: %d\n", x)
			return false
		}
		if !lm.ClientIP.Equal(m.ClientIP) || !lm.YourIP.Equal(m.YourIP) {
			// Fixme: verify more fields.
			fmt.Printf("XFAIL: %+v -> %+v\n", lm.ClientIP, m.ClientIP)
			return false
		}
		if opt.SubnetMask == nil || opt.Routers == nil || opt.ServerIdentifier == nil ||
			m.YourIP.Equal(net.IPv4(0, 0, 0, 0)) || m.YourIP.Equal(net.IPv4(255, 255, 255, 255)) {
			return false
		}
		return true
	}
}

func verifyDiscover(xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) bool {
	return func(m dhcpmsg.Message, opt dhcpmsg.DecodedOptions) bool {
		if m.Xid != xid {
			return false
		}
		if x := opt.MessageType; x == nil || *x != dhcpmsg.MsgTypeOffer {
			return false
		}
		if opt.SubnetMask == nil || opt.Routers == nil || opt.ServerIdentifier == nil ||
			m.YourIP.Equal(net.IPv4(0, 0, 0, 0)) || m.YourIP.Equal(net.IPv4(255, 255, 255, 255)) {
			return false
		}
		return true
	}
}
