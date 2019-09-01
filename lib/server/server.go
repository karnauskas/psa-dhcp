package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
	"gitlab.com/adrian_blx/psa-dhcp/lib/libif"
	"gitlab.com/adrian_blx/psa-dhcp/lib/server/ipdb"
	pb "gitlab.com/adrian_blx/psa-dhcp/lib/server/proto"
)

type server struct {
	ctx       context.Context         // Context used by this server.
	l         *log.Logger             // Logger.
	iface     *net.Interface          // Interface we are working on.
	selfIP    net.IP                  // Our own IP (used as server identifier).
	ipdb      *ipdb.IPDB              // IP database instance.
	lopts     leaseOptions            // Default options for leases.
	overrides map[string]leaseOptions // Static client configuration.
}

type leaseOptions struct {
	ip            net.IP        // Static IP of a lease.
	domain        string        // Domain to announce.
	hostname      string        // Hostname to use.
	netmask       net.IPMask    // Netmask of the network we announce.
	router        net.IP        // Router to use.
	dns           []net.IP      // List of DNS suggested to the client.
	ntp           []net.IP      // List of NTP servers suggested to the client.
	leaseDuration time.Duration // Duration of the announced lease.
}

func New(ctx context.Context, l *log.Logger, iface *net.Interface, conf *pb.ServerConfig) (*server, error) {
	selfIP, err := libif.InterfaceAddr(iface)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch my own IP from interface '%s': %v", iface.Name, err)
	}

	lopts, ipnet, err := parseConfig(conf)
	if err != nil {
		return nil, fmt.Errorf("config parse error: %v", err)
	}

	db, err := ipdb.New(ipnet.IP, ipnet.Mask)
	if err != nil {
		return nil, fmt.Errorf("failed to build ipdb: %v", err)
	}

	// Check dynamic range configuration: It is not part of the lease config but a property of ipdb.
	if dr := conf.GetDynamicRange(); dr != "" {
		sr := strings.Split(dr, "-")
		if len(sr) != 2 {
			return nil, fmt.Errorf("dynamic_range value '%s' invalid. Expected 'start-end'", dr)
		}
		ipa := net.ParseIP(sr[0])
		ipb := net.ParseIP(sr[1])
		if ipa == nil || ipb == nil {
			return nil, fmt.Errorf("dynamic_range has invalid IP range: '%s'", dr)
		}
		if err := db.SetDynamicRange(ipa, ipb); err != nil {
			return nil, fmt.Errorf("failed to configure dynamic_range '%s': %v", dr, err)
		}
		l.Printf("# dynamic range restricted to %s", dr)
	}

	// Disable dynamic ranges if desired
	if conf.GetStaticOnly() {
		db.DisableDynamic()
		l.Printf("# disabling dynamic IP assignment (static_only is 'true'), only static leases will be handed out.")
	}

	overrides := make(map[string]leaseOptions)
	for k, v := range conf.GetClient() {
		hwaddr, err := net.ParseMAC(k)
		if err != nil {
			return nil, fmt.Errorf("failed to parse hwaddr '%s': %v", k, err)
		}
		oopts := *lopts
		setClientOverrides(&oopts, v)
		if oopts.ip != nil {
			if err := db.AddPermanentClient(oopts.ip, duidFromHwAddr(hwaddr)); err != nil {
				return nil, fmt.Errorf("could not create permanent lease for %v -> %v: %v", hwaddr, oopts.ip, err)
			}
		}
		if _, ok := overrides[hwaddr.String()]; ok {
			return nil, fmt.Errorf("duplicate permanent lease for %v", hwaddr)
		}
		overrides[hwaddr.String()] = oopts
	}

	// Give ourselfs a permanent fake lease.
	if err := db.AddPermanentClient(selfIP, duidFromHwAddr(iface.HardwareAddr)); err != nil {
		return nil, fmt.Errorf("failed to add own IP (%s) to configured net (%s): %v", selfIP, *ipnet, err)
	}
	return &server{ctx: ctx, l: l, iface: iface, selfIP: selfIP, ipdb: db, lopts: *lopts, overrides: overrides}, nil
}

// dhcpOptions assembles a list of dhcp options from the server configuration.
func (sx *server) dhcpOptions() []dhcpmsg.DHCPOpt {
	opts := []dhcpmsg.DHCPOpt{
		dhcpmsg.OptionIPAddressLeaseDuration(sx.lopts.leaseDuration),
		dhcpmsg.OptionSubnetMask(sx.lopts.netmask),
	}
	if sx.lopts.router != nil {
		opts = append(opts, dhcpmsg.OptionRouter(sx.lopts.router))
	}
	if len(sx.lopts.dns) > 0 {
		opts = append(opts, dhcpmsg.OptionDNS(sx.lopts.dns...))
	}
	if len(sx.lopts.ntp) > 0 {
		opts = append(opts, dhcpmsg.OptionNTP(sx.lopts.ntp...))
	}
	if sx.lopts.domain != "" {
		opts = append(opts, dhcpmsg.OptionDomainName(sx.lopts.domain))
	}
	return opts
}

func (sx *server) String() string {
	return fmt.Sprintf("server(iface=%s, ip=%s, lease_opts=%+v)", sx.iface.Name, sx.selfIP, sx.lopts)
}
