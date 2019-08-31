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
	"gitlab.com/adrian_blx/psa-dhcp/lib/server/ipdb/uip"
)

type IPDB struct {
	sync.RWMutex
	from    uip.Uip // Lowest IP we manage (probably the network)
	to      uip.Uip // Highest IP we manage (might be bcast)
	clients *clients.Clients
}

func New(network net.IP, netmask net.IPMask) (*IPDB, error) {
	from, to, err := fromTo(network, netmask)
	if err != nil {
		return nil, err
	}
	return &IPDB{from: from, to: to, clients: clients.NewClients()}, nil
}

func (ix *IPDB) LookupClientByHwAddr(hwaddr net.HardwareAddr) (net.IP, error) {
	ix.Lock()
	defer ix.Unlock()

	_, res := ix.clients.Lookup(time.Now(), uip.Uip(0), hwaddr)
	if res == nil {
		return nil, fmt.Errorf("no such client found")
	}
	return res.Uip().ToV4(), nil
}

// AddPermanentClient injects a new client and marks it as permanent.
// While the lease may expire, the ip<>hwaddr mapping will not.
func (ix *IPDB) AddPermanentClient(ip net.IP, hwaddr net.HardwareAddr) error {
	ix.Lock()
	defer ix.Unlock()

	n, err := ix.toUip(ip)
	if err != nil {
		return err
	}
	return ix.clients.InjectPermanent(time.Now(), n, hwaddr)
}

// SetClient updates the state of a client, inserting it if needed.
func (ix *IPDB) UpdateClient(ip net.IP, hwaddr net.HardwareAddr, ttl time.Duration) error {
	ix.Lock()
	defer ix.Unlock()

	n, err := ix.toUip(ip)
	if err != nil {
		return err
	}
	now := time.Now()
	ltime := now.Add(ttl)

	// First, just try an optimistic set.
	if ix.clients.SetLease(now, n, hwaddr, ltime) == nil {
		return nil
	}
	// If this failed, we might need to inject first.
	if err := ix.clients.Inject(now, n, hwaddr, ltime); err != nil {
		return err
	}
	return ix.clients.SetLease(now, n, hwaddr, ltime)
}

// FindIP attempts to find an IP for given hwaddr, having a bias for the suggested IP.
func (ix *IPDB) FindIP(ctx context.Context, isFree func(context.Context, net.IP, net.HardwareAddr) bool, ip net.IP, hwaddr net.HardwareAddr) (net.IP, error) {
	ix.Lock()
	defer ix.Unlock()

	n, err := ix.toUip(ip)
	if err != nil {
		// Suggested IP not in range, just ignore it.
		n = uip.Uip(0)
	}

	oip, ohwaddr := ix.clients.Lookup(time.Now(), n, hwaddr)
	if ohwaddr != nil {
		// This hardware addr already has a lease.
		return ohwaddr.Uip().ToV4(), nil
	}

	p := rand.Perm(1 + int(ix.to-ix.from))
	if oip == nil {
		p = append([]int{int(n - ix.from)}, p...)
	}

	for _, v := range p {
		if ctx.Err() != nil {
			break
		}
		picked := ix.from + uip.Uip(v)
		e, _ := ix.clients.Lookup(time.Now(), picked, nil)
		if e == nil && picked.Valid() && isFree(ctx, picked.ToV4(), hwaddr) {
			return picked.ToV4(), nil
		}
	}
	return nil, fmt.Errorf("no free ip found")
}

func (ix *IPDB) toUip(ip net.IP) (uip.Uip, error) {
	var n uip.Uip
	if v4 := ip.To4(); v4 == nil {
		return 0, fmt.Errorf("not an ipv4")
	} else {
		n = uip.Uip(binary.BigEndian.Uint32(v4))
	}

	if n < ix.from || n > ix.to {
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
