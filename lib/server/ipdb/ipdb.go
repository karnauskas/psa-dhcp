package ipdb

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/server/ipdb/clients"
	d "gitlab.com/adrian_blx/psa-dhcp/lib/server/ipdb/duid"
	"gitlab.com/adrian_blx/psa-dhcp/lib/server/ipdb/uip"
)

type IPDB struct {
	sync.RWMutex
	netFrom uip.Uip // Lowest IP we manage.
	netTo   uip.Uip // Highest IP we manage.
	dynFrom uip.Uip // Lowest IP to hand out while searching for IPs. If dynFrom and dynTo are set to zero, dynamic searches are disabled.
	dynTo   uip.Uip // Highest IP to hand out while searching for IPs.
	clients *clients.Clients
}

func New(network net.IP, netmask net.IPMask) (*IPDB, error) {
	from, to, err := fromTo(network, netmask)
	if err != nil {
		return nil, err
	}
	return &IPDB{
		netFrom: from,
		netTo:   to,
		dynFrom: from,
		dynTo:   to,
		clients: clients.NewClients(),
	}, nil
}

// SetDynamicRange configures the pool range to find dynamic IPs.
func (ix *IPDB) SetDynamicRange(begin, end net.IP) error {
	ix.Lock()
	defer ix.Unlock()

	b, err := ix.toUip(begin)
	if err != nil {
		return err
	}

	e, err := ix.toUip(end)
	if err != nil {
		return err
	}

	if b > e {
		// toUip already checked the network range, so we only need to validate sanity.
		return fmt.Errorf("begin in dynamic range can not be larget than end")
	}
	ix.dynFrom = b
	ix.dynTo = e
	return nil
}

// DisableDynamic configures ipdb to only hand out pre-configured (or still existing) leases.
// No search will be performed for dynamic leases.
func (ix *IPDB) DisableDynamic() {
	ix.Lock()
	defer ix.Unlock()
	ix.dynFrom = 0
	ix.dynTo = 0
}

func (ix *IPDB) LookupClientByDuid(duid d.Duid) (net.IP, error) {
	ix.Lock()
	defer ix.Unlock()

	_, res := ix.clients.Lookup(time.Now(), uip.Uip(0), duid)
	if res == nil {
		return nil, fmt.Errorf("no such client found")
	}
	return res.Uip().ToV4(), nil
}

// AddPermanentClient injects a new client and marks it as permanent.
// While the lease may expire, the ip<>duid mapping will not.
func (ix *IPDB) AddPermanentClient(ip net.IP, duid d.Duid) error {
	ix.Lock()
	defer ix.Unlock()

	n, err := ix.toUip(ip)
	if err != nil {
		return err
	}
	return ix.clients.InjectPermanent(time.Now(), n, duid)
}

// SetClient updates the state of a client, inserting it if needed.
func (ix *IPDB) UpdateClient(ip net.IP, duid d.Duid, ttl time.Duration) error {
	ix.Lock()
	defer ix.Unlock()

	n, err := ix.toUip(ip)
	if err != nil {
		return err
	}
	now := time.Now()
	ltime := now.Add(ttl)

	// First, just try an optimistic set.
	if ix.clients.SetLease(now, n, duid, ltime) == nil {
		return nil
	}
	// If this failed, we might need to inject first.
	if err := ix.clients.Inject(now, n, duid, ltime); err != nil {
		return err
	}
	return ix.clients.SetLease(now, n, duid, ltime)
}

// FindIP attempts to find an IP for given duid, having a bias for the suggested IP.
func (ix *IPDB) FindIP(ctx context.Context, isFree func(context.Context, net.IP) bool, ip net.IP, duid d.Duid) (net.IP, error) {
	ix.Lock()
	defer ix.Unlock()

	n, err := ix.toUip(ip)
	if err != nil {
		// Suggested IP not in range, just ignore it.
		n = uip.Uip(0)
	}

	oip, oduid := ix.clients.Lookup(time.Now(), n, duid)
	if oduid != nil {
		// This duid already has a lease.
		return oduid.Uip().ToV4(), nil
	}

	if ix.dynTo == 0 && ix.dynFrom == 0 {
		return nil, fmt.Errorf("dynamic searches are disabled")
	}

	p := rand.Perm(1 + int(ix.dynTo-ix.dynFrom))
	if oip == nil {
		p = append([]int{int(n - ix.dynFrom)}, p...)
	}

	for _, v := range p {
		if ctx.Err() != nil {
			break
		}
		picked := ix.dynFrom + uip.Uip(v)
		e, _ := ix.clients.Lookup(time.Now(), picked, nil)
		if e == nil && picked.Valid() && isFree(ctx, picked.ToV4()) {
			return picked.ToV4(), nil
		}
	}
	return nil, fmt.Errorf("no free ip found")
}

// InManagedRange returns 'true' if given ip is in the network range we manage.
func (ix *IPDB) InManagedRange(ip net.IP) bool {
	if _, err := ix.toUip(ip); err == nil {
		return true
	}
	return false
}

func (ix *IPDB) toUip(ip net.IP) (uip.Uip, error) {
	var n uip.Uip
	if v4 := ip.To4(); v4 == nil {
		return 0, fmt.Errorf("not an ipv4")
	} else {
		n = uip.Uip(binary.BigEndian.Uint32(v4))
	}

	if n < ix.netFrom || n > ix.netTo {
		return 0, fmt.Errorf("ip is not in managed range")
	}
	return n, nil
}

func fromTo(network net.IP, netmask net.IPMask) (uip.Uip, uip.Uip, error) {
	v4 := network.To4()
	if v4 == nil {
		return 0, 0, fmt.Errorf("invalid network")
	}
	if len(netmask) != 4 {
		return 0, 0, fmt.Errorf("invalid netmask")
	}

	n := binary.BigEndian.Uint32(v4)
	nm := binary.BigEndian.Uint32(netmask)

	start := n & nm
	end := start + ^nm
	if start != end {
		start++
		end--
	}
	return uip.Uip(start), uip.Uip(end), nil
}
