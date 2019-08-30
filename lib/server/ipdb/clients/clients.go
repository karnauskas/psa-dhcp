package clients

import (
	"fmt"
	"net"
	"sync"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/server/ipdb/uip"
)

type client struct {
	ip          uip.Uip          // IP of this client.
	hwaddr      net.HardwareAddr // clients hardware address.
	leasedUntil time.Time        // validity of this lease.
	permanent   bool             // permanent entries expire, but are never removed.
}

type Clients struct {
	sync.RWMutex
	m map[string]*client
}

func NewClients() *Clients {
	return &Clients{m: make(map[string]*client)}
}

// Lookup returns the entries matching given IP and/or hwaddr.
// Permanent clients will never expire and are always returned while other entries will only
// appear until their lease expired.
func (cx *Clients) Lookup(now time.Time, ip uip.Uip, hwaddr net.HardwareAddr) (*client, *client) {
	cx.Lock()
	defer cx.Unlock()

	var res [2]*client

	for i, k := range []string{ip.String(), hwaddr.String()} {
		if p := cx.m[k]; p != nil {
			if !p.permanent && now.After(p.leasedUntil) {
				delete(cx.m, k)
			} else {
				res[i] = p
			}
		}
	}
	return res[0], res[1]
}

func (cx *Clients) InjectPermanent(now time.Time, ip uip.Uip, hwaddr net.HardwareAddr) error {
	return cx.injectInternal(now, ip, hwaddr, time.Unix(0, 0), true)
}

func (cx *Clients) Inject(now time.Time, ip uip.Uip, hwaddr net.HardwareAddr, leasedUntil time.Time) error {
	return cx.injectInternal(now, ip, hwaddr, leasedUntil, false)
}

func (cx *Clients) injectInternal(now time.Time, ip uip.Uip, hwaddr net.HardwareAddr, leasedUntil time.Time, permanent bool) error {
	cx.Lock()
	cx.Unlock()

	if ip, hw := cx.Lookup(now, ip, hwaddr); ip != nil {
		return fmt.Errorf("entry for ip already exists")
	} else if hw != nil {
		return fmt.Errorf("entry for this hardwareaddr already exists")
	}
	c := &client{ip: ip, hwaddr: hwaddr, leasedUntil: leasedUntil, permanent: permanent}
	cx.m[c.ip.String()] = c
	cx.m[c.hwaddr.String()] = c
	return nil
}

func (cx *Clients) SetLease(now time.Time, ip uip.Uip, hwaddr net.HardwareAddr, leasedUntil time.Time) error {
	cx.Lock()
	cx.Unlock()

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

func (cx *Clients) Expire(now time.Time, ip uip.Uip, hwaddr net.HardwareAddr) error {
	return cx.SetLease(now, ip, hwaddr, time.Unix(0, 0))
}

func (c *client) Uip() uip.Uip {
	return c.ip
}
