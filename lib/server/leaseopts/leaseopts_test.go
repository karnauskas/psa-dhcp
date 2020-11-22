package leaseopts

import (
	"net"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	pb "git.sr.ht/~adrian-blx/psa-dhcp/lib/server/proto"
)

func TestParseConfig(t *testing.T) {
	pp := pb.ServerConfig{
		Network:       "192.168.1.0/24",
		LeaseDuration: "1m5s",
		Domain:        "funky",
		Router:        "192.168.1.1",
		Dns:           []string{"192.168.1.1", "192.168.1.2"},
		Ntp:           []string{"192.168.1.8", "192.168.1.9"},
	}

	lopts, ipnet, err := ParseConfig(&pp)
	if err != nil {
		t.Errorf("ParseConfig(#first) had error: %v", err)
	}
	if diff := cmp.Diff(ipnet, &net.IPNet{IP: net.IPv4(192, 168, 1, 0), Mask: net.IPMask{255, 255, 255, 0}}); diff != "" {
		t.Errorf("Diff(#ipnet): %s", diff)
	}
	if diff := cmp.Diff(lopts, &LeaseOptions{
		Domain:        "funky",
		Router:        net.IPv4(192, 168, 1, 1),
		Netmask:       net.IPMask{255, 255, 255, 0},
		DNS:           []net.IP{net.IPv4(192, 168, 1, 1), net.IPv4(192, 168, 1, 2)},
		NTP:           []net.IP{net.IPv4(192, 168, 1, 8), net.IPv4(192, 168, 1, 9)},
		LeaseDuration: 65 * time.Second,
	}); diff != "" {
		t.Errorf("Diff(#lopts): %s", diff)
	}

	if _, _, err := ParseConfig(&pb.ServerConfig{Router: "::1"}); err == nil {
		t.Errorf("ParseConfig(#bad) wanted err, got nil err")
	}
}

func TestClientOverrides(t *testing.T) {
	cc := pb.ClientConfig{
		Ip:       "192.168.1.99",
		Router:   "192.168.1.1",
		Hostname: "funkytown",
		Dns:      []string{"192.168.1.111", "192.168.1.112"},
		Ntp:      []string{"8.8.1.1", "8.8.2.2"},
	}
	lopts := LeaseOptions{}

	if err := SetClientOverrides(&lopts, &cc); err != nil {
		t.Errorf("SetClientOverrides(#first) = %v, wanted no err", err)
	}
	if diff := cmp.Diff(lopts, LeaseOptions{
		IP:       net.IPv4(192, 168, 1, 99),
		Hostname: "funkytown",
		Router:   net.IPv4(192, 168, 1, 1),
		DNS:      []net.IP{net.IPv4(192, 168, 1, 111), net.IPv4(192, 168, 1, 112)},
		NTP:      []net.IP{net.IPv4(8, 8, 1, 1), net.IPv4(8, 8, 2, 2)},
	}); diff != "" {
		t.Errorf("SetClientOverrides(#first) had a diff: %s", diff)
	}

	// No changes expected
	if err := SetClientOverrides(&lopts, &pb.ClientConfig{}); err != nil {
		t.Errorf("SetClientOverrides(#second) = %v, wanted no err", err)
	}
	if diff := cmp.Diff(lopts, LeaseOptions{
		IP:       net.IPv4(192, 168, 1, 99),
		Hostname: "funkytown",
		Router:   net.IPv4(192, 168, 1, 1),
		DNS:      []net.IP{net.IPv4(192, 168, 1, 111), net.IPv4(192, 168, 1, 112)},
		NTP:      []net.IP{net.IPv4(8, 8, 1, 1), net.IPv4(8, 8, 2, 2)},
	}); diff != "" {
		t.Errorf("SetClientOverrides(#second) had a diff: %s", diff)
	}

	// Overwrite some
	if err := SetClientOverrides(&lopts, &pb.ClientConfig{Hostname: "a", Ntp: []string{"127.0.0.1"}}); err != nil {
		t.Errorf("SetClientOverrides(#third) = %v, wanted no err", err)
	}
	if diff := cmp.Diff(lopts, LeaseOptions{
		IP:       net.IPv4(192, 168, 1, 99),
		Hostname: "a",
		Router:   net.IPv4(192, 168, 1, 1),
		DNS:      []net.IP{net.IPv4(192, 168, 1, 111), net.IPv4(192, 168, 1, 112)},
		NTP:      []net.IP{net.IPv4(127, 0, 0, 1)},
	}); diff != "" {
		t.Errorf("SetClientOverrides(#third) had a diff: %s", diff)
	}

	// Bad values
	if err := SetClientOverrides(&lopts, &pb.ClientConfig{Ip: "192.168.1.55", Ntp: []string{"::1"}}); err == nil {
		t.Errorf("SetClientOverrides(#fourth) returned no error, wanted err")
	}
	if diff := cmp.Diff(lopts, LeaseOptions{
		IP:       net.IPv4(192, 168, 1, 99),
		Hostname: "a",
		Router:   net.IPv4(192, 168, 1, 1),
		DNS:      []net.IP{net.IPv4(192, 168, 1, 111), net.IPv4(192, 168, 1, 112)},
		NTP:      []net.IP{net.IPv4(127, 0, 0, 1)},
	}); diff != "" {
		t.Errorf("SetClientOverrides(#third) had a diff: %s", diff)
	}
}

func TestIpv4(t *testing.T) {
	input := []struct {
		name    string
		input   []string
		wantIP  []net.IP
		wantErr bool
	}{
		{
			name:  "empty",
			input: []string{""},
		},
		{
			name:   "one IPv4",
			input:  []string{"192.168.1.1"},
			wantIP: []net.IP{net.IPv4(192, 168, 1, 1)},
		},
		{
			name:   "two IPv4",
			input:  []string{"192.168.1.1", "192.168.4.9"},
			wantIP: []net.IP{net.IPv4(192, 168, 1, 1), net.IPv4(192, 168, 4, 9)},
		},
		{
			name:    "junk",
			input:   []string{"192.168.1.1/24"},
			wantErr: true,
		},
		{
			name:    "ipv6",
			input:   []string{"::1"},
			wantErr: true,
		},
	}

	for _, test := range input {
		ip, err := ipv4(test.input...)
		if test.wantErr && err == nil {
			t.Errorf("ipv4(%s) expected err, got nil", test.name)
		}
		if err != nil && !test.wantErr {
			t.Errorf("ipv4(%s) = %v, wanted nil-err", test.name, err)
		}
		if diff := cmp.Diff(test.wantIP, ip); diff != "" {
			t.Errorf("ipv4(%s) had diff: %s", test.name, diff)
		}
	}
}
