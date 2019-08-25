package ipdb

import (
	"net"
)

type uip uint32

func (ux uip) toV4() net.IP {
	return net.IPv4(byte(ux>>24)&0xFF, byte(ux>>16)&0xFF, byte(ux>>8)&0xFF, byte(ux)&0xFF)
}

func (ux uip) valid() bool {
	return ux&0xFF != 0 && ux&0xFF != 0xFF
}
