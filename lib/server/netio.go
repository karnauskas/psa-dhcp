package server

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
	d "gitlab.com/adrian_blx/psa-dhcp/lib/server/ipdb/duid"
	"gitlab.com/adrian_blx/psa-dhcp/lib/server/replies"
	yl "gitlab.com/adrian_blx/psa-dhcp/lib/server/ylog"
)

func (sx *server) handleMsg(src, dst net.IP, msg dhcpmsg.Message) {
	opts := dhcpmsg.DecodeOptions(msg.Options)
	duid := sx.getDuid(msg.ClientMAC, opts.ClientIdentifier)
	yl := yl.New(sx.l, msg, opts)

	// Some sanity checks before handling this message.
	if bytes.Equal(sx.iface.HardwareAddr, msg.ClientMAC) {
		yl.Printf("received a message with my own hwaddr from duid %s, dropping.", duid)
		return
	}
	if sx.selfIP.Equal(opts.RequestedIP) {
		yl.Printf("received request for my own IP from duid %s, nice try...", duid)
		return
	}

	switch opts.MessageType {
	case dhcpmsg.MsgTypeDiscover:
		// 50% chance of delaying the replay to give 'slower' DHCP servers a chance.
		if delay := time.Duration(rand.Int63n(2)*50) * time.Millisecond; delay > 0 {
			yl.Printf("DISCOVER: Waiting for %v for other servers to pick up.", delay)
			time.Sleep(delay)
		}
		sx.handleDiscover(yl, src, dst, duid, msg, opts)
	case dhcpmsg.MsgTypeRequest:
		sx.handleRequest(yl, src, dst, duid, msg, opts)
	default:
		yl.Printf("dropping unhandled message of type %d", opts.MessageType)
		// ignored
	}
}

func (sx *server) handleDiscover(yl *yl.Ylog, src, dst net.IP, duid d.Duid, msg dhcpmsg.Message, opts dhcpmsg.DecodedOptions) {
	if !dst.Equal(net.IPv4bcast) {
		yl.Printf("DISCOVER: Oops! Client with IP %s sent this to destination %s, should have been broadcasted. Dropping!", src, dst)
		return
	}
	if opts.ServerIdentifier != nil {
		yl.Printf("DISCOVER: Oops! Client with DUID %s specified a server identifier! Dropping!", duid)
		return
	}

	yl.Printf("DISCOVER: Searching for a free IP, client suggested IP '%s'", opts.RequestedIP)
	offer, err := sx.ipdb.FindIP(sx.ctx, sx.arpVerify(msg.ClientMAC), opts.RequestedIP, duid)
	if err != nil {
		yl.Printf("DISCOVER: Failed to find a free IP")
		return
	}
	if err := sx.ipdb.UpdateClient(offer, duid, 15*time.Second); err != nil {
		yl.Printf("DISCOVER: Failed to update temporarily lease during discovery")
		return
	}

	yl.Printf("DISCOVER: Sending offer for IP '%s' to DUID '%s'", offer, duid)
	sx.sendMsg(msg, offer, replies.AssembleOffer)
}

func (sx *server) handleRequest(yl *yl.Ylog, src, dst net.IP, duid d.Duid, msg dhcpmsg.Message, opts dhcpmsg.DecodedOptions) {
	/*
	   ---------------------------------------------------------------------
	   |              |INIT-REBOOT  |SELECTING    |RENEWING     |REBINDING |
	   ---------------------------------------------------------------------
	   |broad/unicast |broadcast    |broadcast    |unicast      |broadcast |
	   |server-ident  |MUST NOT     |MUST         |MUST NOT     |MUST NOT  |
	   |requested-ip  |MUST         |MUST         |MUST NOT     |MUST NOT  |
	   ---------------------------------------------------------------------
	*/

	var desiredIP net.IP
	if dst.Equal(net.IPv4bcast) && opts.ServerIdentifier == nil && opts.RequestedIP != nil {
		// INIT-Reboot
		yl.Printf("REQUEST: INIT-Reboot client desires IP '%s'", opts.RequestedIP)
		desiredIP = opts.RequestedIP
	} else if dst.Equal(net.IPv4bcast) && opts.ServerIdentifier.Equal(sx.selfIP) && opts.RequestedIP != nil {
		// SELECTING
		yl.Printf("REQUEST: SELECTING state for DUID '%s'", duid)
		desiredIP = opts.RequestedIP
	} else if dst.Equal(sx.selfIP) && opts.ServerIdentifier == nil && opts.RequestedIP == nil {
		// RENEWING
		yl.Printf("REQUEST: RENEWAL from IP '%s'", src)
		desiredIP = src
	} else if dst.Equal(net.IPv4bcast) && opts.ServerIdentifier == nil && opts.RequestedIP == nil {
		// REBINDING
		yl.Printf("REQUEST: REBINDING from IP '%s'", src)
		desiredIP = src
	} else {
		yl.Printf("REQUEST: Bogous request for destination '%s' with server identifier '%s' dropped", dst, opts.ServerIdentifier)
		return
	}

	// BUG.
	if desiredIP == nil {
		panic(fmt.Errorf("desiredIP is nil"))
	}

	lease, err := sx.ipdb.LookupClientByDuid(duid)
	if err != nil {
		yl.Printf("REQUEST: Failed to find lease for DUID '%s', sending NAK: %v", duid, err)
		sx.sendNACK(msg.Xid, msg.ClientMAC)
		return
	}
	if !desiredIP.Equal(lease) {
		yl.Printf("REQUEST: Client wanted IP '%s', but got a lease for '%s', sending NAK", desiredIP, lease)
		sx.sendNACK(msg.Xid, msg.ClientMAC)
		return
	}
	if !sx.arpVerify(msg.ClientMAC)(sx.ctx, lease) {
		yl.Printf("REQUEST: Rejecting lease for '%s' as IP failed ARP check, sending NAK", lease)
		sx.sendNACK(msg.Xid, msg.ClientMAC)
		return
	}
	if err := sx.ipdb.UpdateClient(lease, duid, sx.lopts.LeaseDuration); err != nil {
		// Probably a race condition - just drop it.
		yl.Printf("REQUEST: UpdateClient(%s, %s) failed: %v", lease, duid, err)
		return
	}

	yl.Printf("REQUEST: Lease for '%s' confirmed", lease)
	sx.sendMsg(msg, lease, replies.AssembleACK)
}

func (sx *server) sendNACK(xid uint32, mac net.HardwareAddr) {
	pkt := replies.AssembleNACK(xid, sx.selfIP, mac)
	sx.sendUnicast(mac, pkt)
}

func (sx *server) sendMsg(msg dhcpmsg.Message, ip net.IP, f func(uint32, uint16, net.IP, net.IP, net.HardwareAddr, []dhcpmsg.DHCPOpt) []byte) {
	// FIXME: Overrides
	bcast := (msg.Flags & dhcpmsg.FlagBroadcast) != 0
	pkt := f(msg.Xid, msg.Flags, sx.selfIP, ip, msg.ClientMAC, sx.dhcpOptions())

	if bcast {
		sx.l.Printf(">> SENDING AS BROADCASAT")
		sx.sendUnicast(net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, pkt)
	} else {
		sx.sendUnicast(msg.ClientMAC, pkt)
	}
}
