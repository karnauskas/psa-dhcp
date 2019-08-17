package verify

import (
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
)

var (
	ipInvalid = net.IPv4(0, 0, 0, 0)
	ipBcast   = net.IPv4(255, 255, 255, 255)
)

// VerifyOffer verifies that this message was an offer reply with YourIP and a ServerIdentifier.
func VerifyOffer(xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) bool {
	return func(m dhcpmsg.Message, opt dhcpmsg.DecodedOptions) bool {
		if opt.MessageType != dhcpmsg.MsgTypeOffer {
			return false
		}
		// We don't have a server identifier yet, but need one, so make sure we get one.
		if opt.ServerIdentifier.Equal(ipInvalid) ||
			opt.ServerIdentifier.Equal(ipBcast) {
			return false
		}
		return verifyCommon(xid, m, opt)
	}
}

// VerifySelectingAck checks the ACK message sent to a selecting DHCPREQUEST.
func VerifySelectingAck(lm dhcpmsg.Message, lopt dhcpmsg.DecodedOptions, xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) bool {
	return verifyGenAck(lm, lopt, xid)
}

// VerifyRebindingAck checks the ACK message of a DHCPREQUEST while rebinding.
func VerifyRebindingAck(lm dhcpmsg.Message, lopt dhcpmsg.DecodedOptions, xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) bool {
	return verifyGenAck(lm, lopt, xid)
}

// VerifyRenewingAck checks the ACK message of a DHCPREQUEST while renewing.
func VerifyRenewingAck(lm dhcpmsg.Message, lopt dhcpmsg.DecodedOptions, xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) bool {
	return verifyGenAck(lm, lopt, xid)
}

func verifyGenAck(lm dhcpmsg.Message, lopt dhcpmsg.DecodedOptions, xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) bool {
	return func(m dhcpmsg.Message, opt dhcpmsg.DecodedOptions) bool {
		if opt.MessageType != dhcpmsg.MsgTypeAck {
			return false
		}
		if !isStable(m, lm, opt, lopt) {
			return false
		}
		return verifyCommon(xid, m, opt)
	}
}

// isStable ensures that the server changed its mind too badly.
func isStable(m, lm dhcpmsg.Message, opt, lopt dhcpmsg.DecodedOptions) bool {
	if !m.YourIP.Equal(lm.YourIP) {
		return false
	}
	if !m.NextIP.Equal(lm.NextIP) {
		return false
	}
	if !opt.ServerIdentifier.Equal(lopt.ServerIdentifier) {
		return false
	}
	return true
}

func verifyCommon(xid uint32, m dhcpmsg.Message, opt dhcpmsg.DecodedOptions) bool {
	if m.Xid != xid {
		return false
	}
	if len(opt.Routers) == 0 ||
		m.YourIP.Equal(ipInvalid) ||
		m.YourIP.Equal(ipBcast) ||
		m.NextIP.Equal(ipBcast) {
		return false
	}
	if opt.IPAddressLeaseTime < 1*time.Minute { // that would be silly.
		return false
	}
	return true
}
