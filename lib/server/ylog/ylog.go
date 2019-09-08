package ylog

import (
	"fmt"
	"log"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
)

type Ylog struct {
	l *log.Logger
	s string
}

func New(l *log.Logger, msg dhcpmsg.Message, opts dhcpmsg.DecodedOptions) *Ylog {
	s := fmt.Sprintf("[%s] ", msg.ClientMAC)
	return &Ylog{l: l, s: s}
}

func (yl *Ylog) Printf(fmt string, args ...interface{}) {
	yl.l.Printf(yl.s+fmt, args...)
}
