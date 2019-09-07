package server

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
	d "gitlab.com/adrian_blx/psa-dhcp/lib/server/ipdb/duid"
	"gitlab.com/adrian_blx/psa-dhcp/lib/server/replies"
)

func (sx *server) handleMsg(src, dst net.IP, msg dhcpmsg.Message) {
	opts := dhcpmsg.DecodeOptions(msg.Options)
	duid := sx.getDuid(msg.ClientMAC, opts.ClientIdentifier)

	// Some sanity checks before handling this message.
	if bytes.Equal(sx.iface.HardwareAddr, msg.ClientMAC) {
		sx.l.Printf("[%s] received a message with our own hwaddr from duid %s, dropped", msg.ClientMAC, duid)
		return
	}
	if sx.selfIP.Equal(opts.RequestedIP) {
		sx.l.Printf("[%s] received a request for our IP from duid %s, nice try...", msg.ClientMAC, duid)
		return
	}

	switch opts.MessageType {
	case dhcpmsg.MsgTypeDiscover:
		sx.handleDiscover(src, dst, duid, msg, opts)
	case dhcpmsg.MsgTypeRequest:
		sx.handleRequest(src, dst, duid, msg, opts)
	default:
		sx.l.Printf("[%s] sent us an unhandled message type: %d", msg.ClientMAC, opts.MessageType)
		// ignored
	}
}

func (sx *server) handleDiscover(src, dst net.IP, duid d.Duid, msg dhcpmsg.Message, opts dhcpmsg.DecodedOptions) {
	if !dst.Equal(net.IPv4bcast) {
		sx.l.Printf("[%s] DISCOVER: oops! Client (ip=%s) sent this to %s instead of broadcasting - dropping!", msg.ClientMAC, src, dst)
		return
	}
	if opts.ServerIdentifier != nil {
		sx.l.Printf("[%s] DISCOVER: oops! Client with duid %s specified a server identifier - dropping!", msg.ClientMAC, duid)
		return
	}

	sx.l.Printf("[%s] DISCOVER: searching free IP (client suggested '%s')", msg.ClientMAC, opts.RequestedIP)
	offer, err := sx.ipdb.FindIP(sx.ctx, sx.arpVerify(msg.ClientMAC), opts.RequestedIP, duid)
	if err != nil {
		sx.l.Printf("[%s] DISCOVER: failed to find a free IP for client with duid %s during DISCOVER: %v", msg.ClientMAC, duid, err)
		return
	}
	if err := sx.ipdb.UpdateClient(offer, duid, 15*time.Second); err != nil {
		sx.l.Printf("[%s] DISCOVER: failed to update temporarily lease during discover for duid %s: %v", msg.ClientMAC, duid, err)
		return
	}

	sx.l.Printf("[%s] DISCOVER: sending offer for IP %s to %s as reply", msg.ClientMAC, offer, duid)
	sx.sendMsg(msg, offer, replies.AssembleOffer)
}

func (sx *server) handleRequest(src, dst net.IP, duid d.Duid, msg dhcpmsg.Message, opts dhcpmsg.DecodedOptions) {
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
		sx.l.Printf("[%s] REQUEST: INIT-Reboot desires IP %s", msg.ClientMAC, opts.RequestedIP)
		desiredIP = opts.RequestedIP
	} else if dst.Equal(net.IPv4bcast) && opts.ServerIdentifier.Equal(sx.selfIP) && opts.RequestedIP != nil {
		// SELECTING
		sx.l.Printf("[%s] REQUEST: SELECTING state for duid %s", msg.ClientMAC, duid)
		desiredIP = opts.RequestedIP
	} else if !dst.Equal(net.IPv4bcast) && opts.ServerIdentifier == nil && opts.RequestedIP == nil {
		// RENEWING
		sx.l.Printf("[%s] REQUEST: RENEWAL of %s from duid %s", msg.ClientMAC, src, duid)
		desiredIP = src
	} else if dst.Equal(net.IPv4bcast) && opts.ServerIdentifier == nil && opts.RequestedIP == nil {
		// REBINDING
		sx.l.Printf("[%s] REQUEST: REBINDING of %s from duid %s", msg.ClientMAC, src, duid)
		desiredIP = src
	} else {
		sx.l.Printf("[%s] REQUEST: Bogous server identifier %s dropped", msg.ClientMAC, opts.ServerIdentifier)
		return
	}

	// BUG.
	if desiredIP == nil {
		panic(fmt.Errorf("desiredIP is nil"))
	}

	lease, err := sx.ipdb.LookupClientByDuid(duid)
	if err != nil {
		sx.l.Printf("[%s] REQUEST: Failed to find lease for client with duid %s: %v, sending NAK", msg.ClientMAC, duid, err)
		sx.sendNACK(msg.Xid, msg.ClientMAC)
		return
	}
	if !desiredIP.Equal(lease) {
		sx.l.Printf("[%s] REQUEST: Client wanted %s but has %s, sending NAK", msg.ClientMAC, desiredIP, lease)
		sx.sendNACK(msg.Xid, msg.ClientMAC)
		return
	}
	if !sx.arpVerify(msg.ClientMAC)(sx.ctx, lease) {
		sx.l.Printf("[%s] REQUEST: Rejecting lease for %s as IP failed ARP check, sending NAK", msg.ClientMAC, lease)
		sx.sendNACK(msg.Xid, msg.ClientMAC)
		return
	}
	if err := sx.ipdb.UpdateClient(lease, duid, sx.lopts.LeaseDuration); err != nil {
		// Probably a race condition - just drop it.
		sx.l.Printf("[%s] REQUEST: UpdateClient(%s, %s) failed - reservation raced", msg.ClientMAC, lease, duid)
		return
	}

	sx.l.Printf("[%s] REQUEST: confirmed lease for %s to duid %s", msg.ClientMAC, lease, duid)
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
