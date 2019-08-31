package clients

import (
	"fmt"
	"sync"
	"time"

	d "gitlab.com/adrian_blx/psa-dhcp/lib/server/ipdb/duid"
	"gitlab.com/adrian_blx/psa-dhcp/lib/server/ipdb/uip"
)

type client struct {
	ip          uip.Uip   // IP of this client.
	duid        d.Duid    // Client ID
	leasedUntil time.Time // validity of this lease.
	permanent   bool      // permanent entries expire, but are never removed.
}

type Clients struct {
	sync.RWMutex
	m map[string]*client
}

func NewClients() *Clients {
	return &Clients{m: make(map[string]*client)}
}

// Lookup returns the entries matching given IP and/or duid.
// Permanent clients will never expire and are always returned while other entries will only
// appear until their lease expired.
func (cx *Clients) Lookup(now time.Time, ip uip.Uip, duid d.Duid) (*client, *client) {
	cx.Lock()
	defer cx.Unlock()

	var res [2]*client

	for i, k := range []string{ip.String(), duid.String()} {
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

func (cx *Clients) InjectPermanent(now time.Time, ip uip.Uip, duid d.Duid) error {
	return cx.injectInternal(now, ip, duid, time.Unix(0, 0), true)
}

func (cx *Clients) Inject(now time.Time, ip uip.Uip, duid d.Duid, leasedUntil time.Time) error {
	return cx.injectInternal(now, ip, duid, leasedUntil, false)
}

func (cx *Clients) injectInternal(now time.Time, ip uip.Uip, duid d.Duid, leasedUntil time.Time, permanent bool) error {
	cx.Lock()
	cx.Unlock()

	if ip, hw := cx.Lookup(now, ip, duid); ip != nil {
		return fmt.Errorf("entry for ip already exists")
	} else if hw != nil {
		return fmt.Errorf("entry for this hardwareaddr already exists")
	}
	c := &client{ip: ip, duid: duid, leasedUntil: leasedUntil, permanent: permanent}
	cx.m[c.ip.String()] = c
	cx.m[c.duid.String()] = c
	return nil
}

func (cx *Clients) SetLease(now time.Time, ip uip.Uip, duid d.Duid, leasedUntil time.Time) error {
	cx.Lock()
	cx.Unlock()

	if ip, duid := cx.Lookup(now, ip, duid); ip == nil {
		return fmt.Errorf("ip does not exist")
	} else if duid == nil {
		return fmt.Errorf("duid does not exist")
	} else if ip != duid {
		return fmt.Errorf("ip != duid")
	} else {
		ip.leasedUntil = leasedUntil
		return nil
	}
}

func (cx *Clients) Expire(now time.Time, ip uip.Uip, duid d.Duid) error {
	return cx.SetLease(now, ip, duid, time.Unix(0, 0))
}

func (c *client) Uip() uip.Uip {
	return c.ip
}
