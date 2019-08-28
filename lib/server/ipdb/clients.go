package ipdb

import (
	"fmt"
	"net"
	"time"
)

type client struct {
	ip          uip              // IP of this client.
	hwaddr      net.HardwareAddr // clients hardware address.
	leasedUntil time.Time        // validity of this lease.
	permanent   bool             // permanent entries expire, but are never removed.
}

type clients map[string]*client

// Lookup returns the entries matching given IP and/or hwaddr.
// Permanent clients will never expire and are always returned while other entries will only
// appear until their lease expired.
func (cx clients) Lookup(now time.Time, ip uip, hwaddr net.HardwareAddr) (*client, *client) {
	var res [2]*client

	for i, k := range []string{ip.String(), hwaddr.String()} {
		if p := cx[k]; p != nil {
			if !p.permanent && now.After(p.leasedUntil) {
				delete(cx, k)
			} else {
				res[i] = p
			}
		}
	}
	return res[0], res[1]
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

	cx[c.ip.String()] = &c
	cx[c.hwaddr.String()] = &c
	return nil
}
