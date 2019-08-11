package client

import (
	"net"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
)

var (
	ipInvalid = net.IPv4(0, 0, 0, 0)
	ipBcast   = net.IPv4(255, 255, 255, 255)
)

func verifyAck(lm dhcpmsg.Message, xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) bool {
	return func(m dhcpmsg.Message, opt dhcpmsg.DecodedOptions) bool {
		if opt.MessageType != dhcpmsg.MsgTypeAck {
			return false
		}
		if !lm.ClientIP.Equal(m.ClientIP) || !lm.YourIP.Equal(m.YourIP) {
			// Fixme: verify more fields.
			return false
		}
		return verifyCommon()(xid, m, opt)
	}
}

func verifyDiscover(xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) bool {
	return func(m dhcpmsg.Message, opt dhcpmsg.DecodedOptions) bool {
		if opt.MessageType != dhcpmsg.MsgTypeOffer {
			return false
		}
		return verifyCommon()(xid, m, opt)
	}
}

func verifyCommon() func(uint32, dhcpmsg.Message, dhcpmsg.DecodedOptions) bool {
	return func(xid uint32, m dhcpmsg.Message, opt dhcpmsg.DecodedOptions) bool {
		if m.Xid != xid {
			return false
		}
		if _, bits := opt.SubnetMask.Size(); bits == 0 {
			return false
		}
		if opt.ServerIdentifier.Equal(ipInvalid) || len(opt.Routers) == 0 ||
			ipInvalid.Equal(m.YourIP) || ipBcast.Equal(m.YourIP) {
			return false
		}
		return true
	}
}
