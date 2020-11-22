package server

import (
	"context"

	"git.sr.ht/~adrian-blx/psa-dhcp/lib/dhcpmsg"
	"git.sr.ht/~adrian-blx/psa-dhcp/lib/layer"
	"git.sr.ht/~adrian-blx/psa-dhcp/lib/rsocks"
)

func (sx *server) Run() error {
	sx.l.Printf("# psa-dhcpd is ready!")
	sx.l.Printf("# Configuration: %s", sx)

	rsock, err := rsocks.GetIPRecvSock(sx.iface)
	if err != nil {
		return err
	}

	// Ensure that we close the read socket if context is done.
	ctx, cancel := context.WithCancel(sx.ctx)
	defer cancel()
	go func() {
		<-ctx.Done()
		rsock.Close()
	}()

	buf := make([]byte, 4096)
	for {
		nr, err := rsock.Read(buf)
		if err != nil {
			return err
		}
		v4, err := layer.DecodeIPv4(buf[0:nr])
		if err != nil {
			continue
		}
		udp, err := layer.DecodeUDP(v4.Data)
		if err != nil {
			continue
		}
		dhcp, err := dhcpmsg.Decode(udp.Data)
		if err != nil || dhcp.Op != dhcpmsg.OpRequest {
			continue
		}
		go sx.handleMsg(v4.Source, v4.Destination, *dhcp)
	}
	return nil
}
