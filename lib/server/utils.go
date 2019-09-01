package server

import (
	"bytes"
	"context"
	"net"

	"gitlab.com/adrian_blx/psa-dhcp/lib/arpping"
	"gitlab.com/adrian_blx/psa-dhcp/lib/rsocks"
	d "gitlab.com/adrian_blx/psa-dhcp/lib/server/ipdb/duid"
)

// arpVerify returns a function which can be used to arp-ping an IP.
// The IP is considered to be free if we receive no reply or if it matches the given hwaddr.
func (sx *server) arpVerify(hw net.HardwareAddr) func(context.Context, net.IP) bool {
	return func(ctx context.Context, ip net.IP) bool {
		for i := 0; i < 3; i++ {
			v, err := arpping.Ping(ctx, sx.iface, sx.selfIP, ip)
			if err == nil {
				// Consider this to be 'free' if the reported mac matches the client.
				return bytes.Equal(v, hw)
			}
		}
		return true
	}
}

// sendUnicast sends given payload to an hwaddr / ip destination.
func (sx *server) sendUnicast(hwaddr net.HardwareAddr, payload []byte) error {
	ss, err := rsocks.GetUnicastSendSock(sx.iface, hwaddr)
	if err != nil {
		return err
	}
	ss.Write(payload)
	return nil
}

// getDuid returns the duid to use for this client, based on the static assignements config.
func (sx *server) getDuid(hwaddr net.HardwareAddr, cid []byte) d.Duid {
	// This duid would only exist in ipdb for clients with a static lease
	sduid := duidFromHwAddr(hwaddr)
	if _, err := sx.ipdb.LookupClientByDuid(sduid); err == nil {
		return sduid
	}
	if len(cid) < 4 {
		sx.l.Printf("Client %s did not report a client identifier?! Using self generated duid %s", hwaddr, sduid)
		return sduid
	}
	// -> no static lease? -> use client given identifier as duid.
	return d.Duid(cid)
}

// duidFromHwAddr constructs a duid for internal use from a plain hwaddr.
func duidFromHwAddr(hw net.HardwareAddr) d.Duid {
	// 0x0003 = DUID-LL
	// 0x0000 = Reserved/invalid hw type -> this is internal.
	return d.Duid(append([]byte{0x00, 0x03, 0x00, 0x00}, hw...))
}
