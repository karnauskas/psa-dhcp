package arpping

import (
	"context"
	"net"
	"time"

	"git.sr.ht/~adrian-blx/psa-dhcp/lib/layer"
	"git.sr.ht/~adrian-blx/psa-dhcp/lib/rsocks"
)

// Ping sends an arp ping to the given destination. The call returns after a valid reply
// was received, after 200 ms passed or after the context expired - whichever happens first.
func Ping(ctx context.Context, iface *net.Interface, src, dst net.IP) (net.HardwareAddr, error) {
	actx, acancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer acancel()

	go sendARPPing(actx, iface, src, dst)
	return catchARPReply(actx, iface, dst)
}

func catchARPReply(octx context.Context, iface *net.Interface, target net.IP) (net.HardwareAddr, error) {
	rs, err := rsocks.GetARPRecvSock(iface)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(octx)
	defer cancel() // ensures to close the socket as soon as we return.

	go func() {
		<-ctx.Done()
		rs.Close()
	}()

	buf := make([]byte, 28)
	for {
		nr, err := rs.Read(buf)
		if err != nil {
			return nil, err
		}
		if arp, err := layer.DecodeARP(buf[0:nr]); err == nil && target.Equal(arp.SenderIP) {
			return arp.SenderMAC, nil
		}
	}
}

func sendARPPing(ctx context.Context, iface *net.Interface, src, dst net.IP) {
	ss, err := rsocks.GetARPSendSock(iface)
	if err != nil {
		return
	}
	defer ss.Close()

	a := layer.ARP{
		Opcode:    layer.ARPOpRequest,
		SenderMAC: iface.HardwareAddr,
		SenderIP:  src,
		TargetMAC: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		TargetIP:  dst,
	}
	for {
		ss.Write(a.Assemble())
		select {
		case <-time.After(time.Second):
			continue
		case <-ctx.Done():
			return
		}
	}
}
