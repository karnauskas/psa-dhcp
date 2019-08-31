package uip

import (
	"fmt"
	"net"
)

type Uip uint32

func (ux Uip) ToV4() net.IP {
	return net.IPv4(byte(ux>>24)&0xFF, byte(ux>>16)&0xFF, byte(ux>>8)&0xFF, byte(ux)&0xFF)
}

func (ux Uip) Valid() bool {
	return ux&0xFF != 0 && ux&0xFF != 0xFF
}

func (ux Uip) String() string {
	return fmt.Sprintf("uip(%x)", uint32(ux))
}
