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

func (cx clients) InjectPermanent(now time.Time, ip uip, hwaddr net.HardwareAddr) error {
	return cx.injectInternal(now, ip, hwaddr, time.Unix(0, 0), true)
}

func (cx clients) Inject(now time.Time, ip uip, hwaddr net.HardwareAddr, leasedUntil time.Time) error {
	return cx.injectInternal(now, ip, hwaddr, leasedUntil, false)
}

func (cx clients) injectInternal(now time.Time, ip uip, hwaddr net.HardwareAddr, leasedUntil time.Time, permanent bool) error {
	if ip, hw := cx.Lookup(now, ip, hwaddr); ip != nil {
		return fmt.Errorf("entry for ip already exists")
	} else if hw != nil {
		return fmt.Errorf("entry for this hardwareaddr already exists")
	}
	c := &client{ip: ip, hwaddr: hwaddr, leasedUntil: leasedUntil, permanent: permanent}
	cx[c.ip.String()] = c
	cx[c.hwaddr.String()] = c
	return nil
}

func (cx clients) SetLease(now time.Time, ip uip, hwaddr net.HardwareAddr, leasedUntil time.Time) error {
	if ip, hw := cx.Lookup(now, ip, hwaddr); ip == nil {
		return fmt.Errorf("ip does not exist")
	} else if hw == nil {
		return fmt.Errorf("hwaddr does not exist")
	} else if ip != hw {
		return fmt.Errorf("ip != hwaddr")
	} else {
		ip.leasedUntil = leasedUntil
		return nil
	}
}

func (cx clients) Expire(now time.Time, ip uip, hwaddr net.HardwareAddr) error {
	return cx.SetLease(now, ip, hwaddr, time.Unix(0, 0))
}
