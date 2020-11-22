package server

import (
	"context"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"git.sr.ht/~adrian-blx/psa-dhcp/lib/dhcpmsg"
	pb "git.sr.ht/~adrian-blx/psa-dhcp/lib/server/proto"
)

func TestServer(t *testing.T) {
	iface, err := net.InterfaceByName("lo")
	if err != nil {
		t.Errorf("setup for lo failed: %v", err)
	}
	l := log.New(os.Stdout, "testing: ", 0)

	conf := &pb.ServerConfig{
		Network:       "127.0.0.1/16",
		LeaseDuration: "5m",
		Domain:        "main",
		Router:        "127.0.0.1",
		Dns:           []string{"192.168.1.2", "192.168.1.3"},
		Ntp:           []string{"192.168.1.4", "192.168.1.5"},
		Client: map[string]*pb.ClientConfig{
			"01:00:00:00:00:00": &pb.ClientConfig{
				Router: "192.168.2.1",
				Dns:    []string{"192.168.2.2", "192.168.2.3"},
				Ntp:    []string{"192.168.2.4", "192.168.2.5"},
			},
			"02:00:00:00:00:00": &pb.ClientConfig{
				Ntp: []string{"192.168.2.4", "192.168.2.5"},
			},
			"03:00:00:00:00:00": &pb.ClientConfig{
				Dns: []string{"192.168.2.2", "192.168.2.3"},
			},
			"04:00:00:00:00:00": &pb.ClientConfig{
				Ip: "127.0.2.1",
			},
		},
	}
	sx, err := New(context.Background(), l, iface, conf)
	if err != nil {
		t.Errorf("New server failed: %v", err)
	}

	input := []struct {
		client string
		want   []dhcpmsg.DHCPOpt
	}{
		{
			client: "01:00:00:00:00:00",
			want: []dhcpmsg.DHCPOpt{
				dhcpmsg.OptionIPAddressLeaseDuration(5 * time.Minute),
				dhcpmsg.OptionSubnetMask(net.IPMask{0xff, 0xff, 0, 0}),
				dhcpmsg.OptionRouter(net.IPv4(192, 168, 2, 1)),
				dhcpmsg.OptionDNS(net.IPv4(192, 168, 2, 2), net.IPv4(192, 168, 2, 3)),
				dhcpmsg.OptionNTP(net.IPv4(192, 168, 2, 4), net.IPv4(192, 168, 2, 5)),
				dhcpmsg.OptionDomainName("main"),
			},
		},
		{
			client: "02:00:00:00:00:00",
			want: []dhcpmsg.DHCPOpt{
				dhcpmsg.OptionIPAddressLeaseDuration(5 * time.Minute),
				dhcpmsg.OptionSubnetMask(net.IPMask{0xff, 0xff, 0, 0}),
				dhcpmsg.OptionRouter(net.IPv4(127, 0, 0, 1)),
				dhcpmsg.OptionDNS(net.IPv4(192, 168, 1, 2), net.IPv4(192, 168, 1, 3)),
				dhcpmsg.OptionNTP(net.IPv4(192, 168, 2, 4), net.IPv4(192, 168, 2, 5)),
				dhcpmsg.OptionDomainName("main"),
			},
		},
		{
			client: "03:00:00:00:00:00",
			want: []dhcpmsg.DHCPOpt{
				dhcpmsg.OptionIPAddressLeaseDuration(5 * time.Minute),
				dhcpmsg.OptionSubnetMask(net.IPMask{0xff, 0xff, 0, 0}),
				dhcpmsg.OptionRouter(net.IPv4(127, 0, 0, 1)),
				dhcpmsg.OptionDNS(net.IPv4(192, 168, 2, 2), net.IPv4(192, 168, 2, 3)),
				dhcpmsg.OptionNTP(net.IPv4(192, 168, 1, 4), net.IPv4(192, 168, 1, 5)),
				dhcpmsg.OptionDomainName("main"),
			},
		},
		{
			client: "04:00:00:00:00:00",
			want: []dhcpmsg.DHCPOpt{
				dhcpmsg.OptionIPAddressLeaseDuration(5 * time.Minute),
				dhcpmsg.OptionSubnetMask(net.IPMask{0xff, 0xff, 0, 0}),
				dhcpmsg.OptionRouter(net.IPv4(127, 0, 0, 1)),
				dhcpmsg.OptionDNS(net.IPv4(192, 168, 1, 2), net.IPv4(192, 168, 1, 3)),
				dhcpmsg.OptionNTP(net.IPv4(192, 168, 1, 4), net.IPv4(192, 168, 1, 5)),
				dhcpmsg.OptionDomainName("main"),
			},
		},
	}
	for _, test := range input {
		mac, err := net.ParseMAC(test.client)
		if err != nil {
			t.Errorf("ParseMAC(%s) = %v; want nil", test.client, err)
		}
		msg := sx.dhcpOptions(mac)
		if diff := cmp.Diff(msg, test.want); diff != "" {
			t.Errorf("Test(%s) failed with diff: %s", mac, diff)
		}
	}
}
