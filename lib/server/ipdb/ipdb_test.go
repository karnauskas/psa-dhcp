package ipdb

import (
	"context"
	"net"
	"testing"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/server/ipdb/uip"
)

func TestOperations(t *testing.T) {
	db, err := New(net.IPv4(192, 168, 2, 0), net.IPv4Mask(255, 255, 255, 0))
	if err != nil {
		t.Fatalf("Could not create ipdb: %v", err)
	}

	if err := db.AddPermanentClient(net.IPv4(192, 168, 2, 1), net.HardwareAddr{0x1}); err != nil {
		t.Errorf("AddPermanentClient(192.168.2.1) failed with %v", err)
	}
	// Both IP and MAC are now reserved
	if err := db.UpdateClient(net.IPv4(192, 168, 2, 1), net.HardwareAddr{0x99}, 1*time.Minute); err == nil {
		t.Errorf("UpdateClient(#wrong mac) did NOT fail")
	}
	if err := db.UpdateClient(net.IPv4(192, 168, 2, 99), net.HardwareAddr{0x1}, 1*time.Minute); err == nil {
		t.Errorf("UpdateClient(#wrong ip) did NOT fail")
	}

	// But we should be able to inject other clients
	if err := db.UpdateClient(net.IPv4(192, 168, 2, 33), net.HardwareAddr{0x33}, 1*time.Minute); err != nil {
		t.Errorf("UpdateClient(#new client) failed with %v", err)
	}
	// Updating the old client should also work.
	if err := db.UpdateClient(net.IPv4(192, 168, 2, 1), net.HardwareAddr{0x1}, 1*time.Minute); err != nil {
		t.Errorf("UpdateClient(#permanent client) failed with %v", err)
	}

	if _, err := db.LookupClientByHwAddr(net.HardwareAddr{0x21}); err == nil {
		t.Errorf("LookupClientByHwAddr(#invalid mac) did NOT fail")
	}
	if res, err := db.LookupClientByHwAddr(net.HardwareAddr{0x33}); err != nil || !res.Equal(net.IPv4(192, 168, 2, 33)) {
		t.Errorf("LookupClientByHwAddr(#existing) failed or returned wrong IP. error=%v, ip=%v", err, res)
	}
}

func TestFindIP(t *testing.T) {
	// Range is limited to 192.168.0.1 - 192.168.0.6
	db, err := New(net.IPv4(192, 168, 0, 1), net.IPv4Mask(255, 255, 255, 248))
	if err != nil {
		t.Fatalf("failed to create ipdb: %v", err)
	}
	ctx := context.Background()

	ip0 := net.IPv4(0, 0, 0, 0)
	ip2 := net.IPv4(192, 168, 0, 2)
	ip3 := net.IPv4(192, 168, 0, 3)
	ip4 := net.IPv4(192, 168, 0, 4)
	isFree := func(ctx context.Context, ip net.IP, hwaddr net.HardwareAddr) bool {
		return !ip.Equal(ip4)
	}

	db.AddPermanentClient(ip2, net.HardwareAddr{0x2})
	db.UpdateClient(ip3, net.HardwareAddr{0x3}, 5*time.Minute)

	// Permanent client with matching IP.
	if ip, err := db.FindIP(ctx, isFree, ip2, net.HardwareAddr{0x2}); err != nil || !ip.Equal(ip2) {
		t.Errorf("FindIP(#permanent1) had error or returned wrong IP. err=%v, ip=%v", err, ip)
	}
	// Non matching IP must still return the permanent entry.
	if ip, err := db.FindIP(ctx, isFree, ip0, net.HardwareAddr{0x2}); err != nil || !ip.Equal(ip2) {
		t.Errorf("FindIP(#permanent2) had error or returned wrong IP. err=%v, ip=%v", err, ip)
	}

	// hwaddr 0x03 should ALWAYS get ip3 as it has a lease
	for i := 0; i < 30; i++ {
		if ip, err := db.FindIP(ctx, isFree, ip0, net.HardwareAddr{0x3}); err != nil || !ip.Equal(ip3) {
			t.Errorf("FindIP(#lease ip3) returned wrong result: err=%v, ip=%v", err, ip)
		}
	}

	// Random client should never get a leased IP or ip4, which is considered non-free.
	for i := 0; i < 30; i++ {
		if ip, err := db.FindIP(ctx, isFree, ip0, net.HardwareAddr{0x99}); err != nil || ip.Equal(ip2) || ip.Equal(ip3) || ip.Equal(ip4) {
			t.Errorf("FindIP(#unleased) returned wrong result: err=%v, ip=%v", err, ip)
		} else {
			u, err := db.toUip(ip)
			if err != nil {
				t.Errorf("could not convert %s to uip: %v", ip, err)
			}
			if u < uip.Uip(0xC0A80001) {
				t.Errorf("FindIP(#unleased) returned ip below our range: %s", u)
			}
			if u > uip.Uip(0xC0A80006) {
				t.Errorf("FindIP(#unleased) returned ip above our range: %s", u)
			}
		}
	}
}

func TestToUip(t *testing.T) {
	db, err := New(net.IPv4(10, 0, 0, 0), net.IPv4Mask(255, 0, 0, 0))
	if err != nil {
		t.Fatalf("Could not create ipdb: %v", err)
	}

	input := []struct {
		name    string
		ip      net.IP
		wantUip uip.Uip
		wantErr bool
	}{
		{
			name:    "base",
			ip:      net.IPv4(10, 0, 0, 0),
			wantErr: true,
		},
		{
			name:    "end",
			ip:      net.IPv4(10, 255, 255, 255),
			wantErr: true,
		},
		{
			name:    "first",
			ip:      net.IPv4(10, 0, 0, 1),
			wantUip: uip.Uip(0x0a000001),
		},
		{
			name:    "last",
			ip:      net.IPv4(10, 255, 255, 254),
			wantUip: uip.Uip(0x0afffffe),
		},
		{
			name:    "IPv6",
			ip:      net.ParseIP("::1"),
			wantErr: true,
		},
	}

	for _, test := range input {
		u, err := db.toUip(test.ip)
		if test.wantErr && err == nil {
			t.Errorf("TestToUip(%s) wanted err, got nil", test.name)
		}
		if err != nil && !test.wantErr {
			t.Errorf("TestToUip(%s) = %v, wanted nil err", test.name, err)
		}
		if u != test.wantUip {
			t.Errorf("TestToUip(%s) = %s, wanted %s", test.name, u, test.wantUip)
		}
	}
}

func TestFromTo(t *testing.T) {
	input := []struct {
		name     string
		network  net.IP
		netmask  net.IPMask
		wantFrom uip.Uip
		wantTo   uip.Uip
	}{
		{
			name:     "zero C",
			network:  net.IPv4(192, 168, 1, 0),
			netmask:  net.IPv4Mask(255, 255, 255, 0),
			wantFrom: uip.Uip(0xc0a80101),
			wantTo:   uip.Uip(0xc0a801fe),
		},
		{
			name:     "normal C",
			network:  net.IPv4(192, 168, 1, 9),
			netmask:  net.IPv4Mask(255, 255, 255, 0),
			wantFrom: uip.Uip(0xc0a80101),
			wantTo:   uip.Uip(0xc0a801fe),
		},
		{
			name:     "none",
			network:  net.IPv4(192, 168, 1, 1),
			netmask:  net.IPv4Mask(255, 255, 255, 255),
			wantFrom: uip.Uip(0xc0a80101),
			wantTo:   uip.Uip(0xc0a80101),
		},
		{
			name:     "all",
			network:  net.IPv4(192, 168, 1, 1),
			netmask:  net.IPv4Mask(0, 0, 0, 0),
			wantFrom: uip.Uip(0x00000001),
			wantTo:   uip.Uip(0xfffffffe),
		},
		{
			name:     "small c",
			network:  net.IPv4(192, 3, 8, 197),
			netmask:  net.IPv4Mask(255, 255, 255, 224),
			wantFrom: uip.Uip(0xC00308C1),
			wantTo:   uip.Uip(0xC00308DE),
		},
		{
			name:     "a class",
			network:  net.IPv4(10, 230, 231, 232),
			netmask:  net.IPv4Mask(255, 240, 0, 0),
			wantFrom: uip.Uip(0x0AE00001),
			wantTo:   uip.Uip(0x0AEFFFFE),
		},
	}

	for _, test := range input {
		from, to, err := fromTo(test.network, test.netmask)
		if err != nil {
			t.Errorf("fromTo(%s) had error %v, wanted nil", test.name, err)
		}
		if from != test.wantFrom {
			t.Errorf("fromTo(%s) from = %s, wanted = %s", test.name, from, test.wantFrom)
		}
		if to != test.wantTo {
			t.Errorf("fromTo(%s) to = %s, wanted = %s", test.name, to, test.wantTo)
		}
	}
}
