package verify

import (
	"net"
	"time"

	"git.sr.ht/~adrian-blx/psa-dhcp/lib/dhcpmsg"
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
		return verifyCommon(xid, m, opt)
	}
}

// VerifySelectingAck checks the ACK message sent to a selecting DHCPREQUEST.
func VerifySelectingAck(lm dhcpmsg.Message, lopt dhcpmsg.DecodedOptions, xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) State {
	return verifyGenAck(lm, lopt, xid, true)
}

// VerifyRenewingAck checks the ACK message of a DHCPREQUEST while renewing.
func VerifyRenewingAck(lm dhcpmsg.Message, lopt dhcpmsg.DecodedOptions, xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) State {
	return verifyGenAck(lm, lopt, xid, true)
}

// VerifyRebindingAck checks the ACK message of a DHCPREQUEST while rebinding.
func VerifyRebindingAck(lm dhcpmsg.Message, lopt dhcpmsg.DecodedOptions, xid uint32) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) State {
	// This is special: We do not care about the server identifier and will take whatever we get.
	return verifyGenAck(lm, lopt, xid, false)
}

func verifyGenAck(lm dhcpmsg.Message, lopt dhcpmsg.DecodedOptions, xid uint32, vrfysi bool) func(dhcpmsg.Message, dhcpmsg.DecodedOptions) State {
	return func(m dhcpmsg.Message, opt dhcpmsg.DecodedOptions) State {
		if opt.MessageType == dhcpmsg.MsgTypeNack {
			return IsNack
		}
		if opt.MessageType != dhcpmsg.MsgTypeAck {
			return Failed
		}
		if !m.YourIP.Equal(lm.YourIP) {
			return Failed
		}
		if vrfysi && !opt.ServerIdentifier.Equal(lopt.ServerIdentifier) {
			return Failed
		}
		return verifyCommon(xid, m, opt)
	}
}

func verifyCommon(xid uint32, m dhcpmsg.Message, opt dhcpmsg.DecodedOptions) State {
	if m.Xid != xid {
		return Failed
	}
	if len(opt.Routers) == 0 ||
		m.YourIP == nil ||
		m.YourIP.Equal(net.IPv4zero) ||
		m.YourIP.Equal(net.IPv4bcast) {
		return Failed
	}
	if opt.ServerIdentifier == nil ||
		opt.ServerIdentifier.Equal(net.IPv4zero) ||
		opt.ServerIdentifier.Equal(net.IPv4bcast) {
		return Failed
	}
	if opt.IPAddressLeaseDuration < 1*time.Minute { // that would be silly.
		return Failed
	}
	return Passed
}
