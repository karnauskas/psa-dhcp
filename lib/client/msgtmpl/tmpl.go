package msgtmpl

import (
	"math/rand"
	"net"
	"os"
	"time"
)

type tmpl struct {
	hostname string
	start    time.Time
	xid      uint32
	hwaddr   [6]byte
}

func New(iface *net.Interface) tmpl {
	t := tmpl{
		start:    time.Now(),
		xid:      rand.Uint32(),
		hostname: "unknown",
	}
	if hn, err := os.Hostname(); err == nil {
		t.hostname = hn
	}
	if len(iface.HardwareAddr) == 6 {
		copy(t.hwaddr[:], iface.HardwareAddr[:])
	}
	return t
}
