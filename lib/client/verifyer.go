package client

import (
	"net"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
)

var (
	ipInvalid = net.IPv4(0, 0, 0, 0)
	ipBcast   = net.IPv4(255, 255, 255, 255)
)

// verifySelectingAck checks the ACK message sent to a selecting DHCPREQUEST.
// We ensure that ClientIP, YourIP and NextIP are still set to the same value as in the previously received message.
func verifySelectingAck(lm dhcpmsg.Message, xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) bool {
	return func(m dhcpmsg.Message, opt dhcpmsg.DecodedOptions) bool {
		if opt.MessageType != dhcpmsg.MsgTypeAck {
			return false
		}
		if !lm.ClientIP.Equal(m.ClientIP) {
			return false
		}
		if !lm.YourIP.Equal(m.YourIP) {
			return false
		}
		if !lm.NextIP.Equal(m.NextIP) {
			return false
		}
		// Options may differ, so we don't check them here.
		return verifyCommon()(xid, m, opt)
	}
}

func verifyRebindingAck(lm dhcpmsg.Message, xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) bool {
	return func(m dhcpmsg.Message, opt dhcpmsg.DecodedOptions) bool {
		if opt.MessageType != dhcpmsg.MsgTypeAck {
			return false
		}
		if !lm.YourIP.Equal(m.YourIP) {
			return false
		}
		if !lm.NextIP.Equal(m.NextIP) {
			return false
		}
		return verifyCommon()(xid, m, opt)
	}
}

func verifyRenewAck(lm dhcpmsg.Message, xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) bool {
	return func(m dhcpmsg.Message, opt dhcpmsg.DecodedOptions) bool {
		if opt.MessageType != dhcpmsg.MsgTypeAck {
			return false
		}
		if !lm.YourIP.Equal(m.YourIP) {
			return false
		}
		if !lm.NextIP.Equal(m.NextIP) {
			return false
		}
		// yes. these should match - but why?
		if !m.ClientIP.Equal(m.YourIP) {
			return false
		}
		return verifyCommon()(xid, m, opt)
	}
}

// verifyOffer verifies that this message was an offer reply with YourIP, NextIP are set.
func verifyOffer(xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) bool {
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
		if len(opt.Routers) == 0 ||
			m.YourIP.Equal(ipInvalid) ||
			m.YourIP.Equal(ipBcast) ||
			m.NextIP.Equal(ipInvalid) ||
			m.NextIP.Equal(ipBcast) {
			return false
		}
		return true
	}
}
