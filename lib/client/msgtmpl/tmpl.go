package msgtmpl

import (
	"net"
	"os"
	"time"
)

type tmpl struct {
	hostname string
	start    time.Time
	xid      uint32
	lastSecs uint16
	hwaddr   [6]byte
}

func New(iface *net.Interface, xid uint32) tmpl {
	t := tmpl{
		start:    time.Now(),
		hostname: "unknown",
		xid:      xid,
	}
	if hn, err := os.Hostname(); err == nil {
		t.hostname = hn
	}
	if len(iface.HardwareAddr) == 6 {
		copy(t.hwaddr[:], iface.HardwareAddr[:])
	}
	return t
}
