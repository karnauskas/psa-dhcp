package ipdb

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

type client struct {
	ip          uip              // uip of this client.
	hwaddr      net.HardwareAddr // clients hardware address.
	leasedUntil time.Time        // validity of this lease.
	permanent   bool             // permanent entries expire, but are never removed.
}

type clients map[uip]client

// Lookup returns the entries matching given IP and/or hwaddr.
func (cx clients) Lookup(now time.Time, ip uip, hwaddr net.HardwareAddr) (*client, *client) {
	var oldIp *client
	var oldHwaddr *client
	for k, v := range cx {
		if !v.permanent && now.After(v.leasedUntil) {
			delete(cx, k)
			continue
		}
		if v.ip == ip {
			cpy := v
			oldIp = &cpy
		}
		if bytes.Equal(v.hwaddr, hwaddr) {
			cpy := v
			oldHwaddr = &cpy
		}
		if oldIp != nil && oldHwaddr != nil {
			break
		}
	}
	return oldIp, oldHwaddr
}

// Inject adds a new client if it doesn't already exist.
func (cx clients) Inject(now time.Time, c client) error {
	ip, hw := cx.Lookup(now, c.ip, c.hwaddr)
	if ip != nil {
		return fmt.Errorf("entry for ip already exists")
	}
	if hw != nil {
		return fmt.Errorf("entry for this hardwareaddr already exists")
	}
	cx[c.ip] = c
	return nil
}
