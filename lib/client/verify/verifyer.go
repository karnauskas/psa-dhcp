package verify

import (
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
)

type State int

const (
	Failed = iota
	Passed
	IsNack
)

// VerifyOffer verifies that this message was an offer reply with YourIP and a ServerIdentifier.
func VerifyOffer(xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) State {
	return func(m dhcpmsg.Message, opt dhcpmsg.DecodedOptions) State {
		if opt.MessageType != dhcpmsg.MsgTypeOffer {
			return Failed
		}
		// We don't have a server identifier yet, but need one, so make sure we get one.
		if opt.ServerIdentifier.Equal(net.IPv4zero) ||
			opt.ServerIdentifier.Equal(net.IPv4bcast) {
			return Failed
		}
		return verifyCommon(xid, m, opt)
	}
}

// VerifySelectingAck checks the ACK message sent to a selecting DHCPREQUEST.
func VerifySelectingAck(lm dhcpmsg.Message, lopt dhcpmsg.DecodedOptions, xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) State {
	return verifyGenAck(lm, lopt, xid)
}

// VerifyRebindingAck checks the ACK message of a DHCPREQUEST while rebinding.
func VerifyRebindingAck(lm dhcpmsg.Message, lopt dhcpmsg.DecodedOptions, xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) State {
	return verifyGenAck(lm, lopt, xid)
}

// VerifyRenewingAck checks the ACK message of a DHCPREQUEST while renewing.
func VerifyRenewingAck(lm dhcpmsg.Message, lopt dhcpmsg.DecodedOptions, xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) State {
	return verifyGenAck(lm, lopt, xid)
}

func verifyGenAck(lm dhcpmsg.Message, lopt dhcpmsg.DecodedOptions, xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) State {
	return func(m dhcpmsg.Message, opt dhcpmsg.DecodedOptions) State {
		if opt.MessageType == dhcpmsg.MsgTypeNack && opt.ServerIdentifier.Equal(lopt.ServerIdentifier) {
			return IsNack
		}
		if opt.MessageType != dhcpmsg.MsgTypeAck {
			return Failed
		}
		if !isStable(m, lm, opt, lopt) {
			return Failed
		}
		return verifyCommon(xid, m, opt)
	}
}

// isStable ensures that the server changed its mind too badly.
func isStable(m, lm dhcpmsg.Message, opt, lopt dhcpmsg.DecodedOptions) bool {
	if !m.YourIP.Equal(lm.YourIP) {
		return false
	}
	if !opt.ServerIdentifier.Equal(lopt.ServerIdentifier) {
		return false
	}
	return true
}

func verifyCommon(xid uint32, m dhcpmsg.Message, opt dhcpmsg.DecodedOptions) State {
	if m.Xid != xid {
		return Failed
	}
	if len(opt.Routers) == 0 ||
		m.YourIP.Equal(net.IPv4zero) ||
		m.YourIP.Equal(net.IPv4bcast) {
		return Failed
	}
	if opt.IPAddressLeaseDuration < 1*time.Minute { // that would be silly.
		return Failed
	}
	return Passed
}
