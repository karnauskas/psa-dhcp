package ylog

import (
	"fmt"
	"log"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
	"gitlab.com/adrian_blx/psa-dhcp/lib/oui"
)

type Ylog struct {
	l *log.Logger
	s string
}

func New(l *log.Logger, msg dhcpmsg.Message, opts dhcpmsg.DecodedOptions) *Ylog {
	vid := "UNKNOWN VENDOR"
	if res, ok := oui.Lookup(msg.ClientMAC); ok {
		vid = res
	}
	s := fmt.Sprintf("[%s] <%-18s> ", msg.ClientMAC, vid)
	return &Ylog{l: l, s: s}
}

func (yl *Ylog) Printf(fmt string, args ...interface{}) {
	yl.l.Printf(yl.s+fmt, args...)
}
