package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
	"gitlab.com/adrian_blx/psa-dhcp/lib/libif"
	"gitlab.com/adrian_blx/psa-dhcp/lib/server/ipdb"
	lo "gitlab.com/adrian_blx/psa-dhcp/lib/server/leaseopts"
	pb "gitlab.com/adrian_blx/psa-dhcp/lib/server/proto"
)

type server struct {
	ctx       context.Context            // Context used by this server.
	l         *log.Logger                // Logger.
	iface     *net.Interface             // Interface we are working on.
	selfIP    net.IP                     // Our own IP (used as server identifier).
	ipdb      *ipdb.IPDB                 // IP database instance.
	lopts     lo.LeaseOptions            // Default options for leases.
	overrides map[string]lo.LeaseOptions // Static client configuration, key is a private duid.
}

// New constructs a new dhcp server instance.
func New(ctx context.Context, l *log.Logger, iface *net.Interface, conf *pb.ServerConfig) (*server, error) {
	selfIP, err := libif.InterfaceAddr(iface)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch my own IP from interface '%s': %v", iface.Name, err)
	}

	lopts, ipnet, err := lo.ParseConfig(conf)
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

	// Configure static assignments
	overrides := make(map[string]lo.LeaseOptions)
	for k, v := range conf.GetClient() {
		hwaddr, err := net.ParseMAC(k)
		if err != nil {
			return nil, fmt.Errorf("failed to parse hwaddr '%s': %v", k, err)
		}
		oopts := *lopts
		lo.SetClientOverrides(&oopts, v)
		if oopts.IP != nil {
			if err := db.AddPermanentClient(oopts.IP, duidFromHwAddr(hwaddr)); err != nil {
				return nil, fmt.Errorf("could not create permanent lease for %v -> %v: %v", hwaddr, oopts.IP, err)
			}
		}
		if _, ok := overrides[hwaddr.String()]; ok {
			return nil, fmt.Errorf("duplicate permanent lease for %v", hwaddr)
		}
		l.Printf("# static mapping for %s configured.", hwaddr)
		overrides[duidFromHwAddr(hwaddr).String()] = oopts
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
		dhcpmsg.OptionIPAddressLeaseDuration(sx.lopts.LeaseDuration),
		dhcpmsg.OptionSubnetMask(sx.lopts.Netmask),
	}
	if sx.lopts.Router != nil {
		opts = append(opts, dhcpmsg.OptionRouter(sx.lopts.Router))
	}
	if len(sx.lopts.DNS) > 0 {
		opts = append(opts, dhcpmsg.OptionDNS(sx.lopts.DNS...))
	}
	if len(sx.lopts.NTP) > 0 {
		opts = append(opts, dhcpmsg.OptionNTP(sx.lopts.NTP...))
	}
	if sx.lopts.Domain != "" {
		opts = append(opts, dhcpmsg.OptionDomainName(sx.lopts.Domain))
	}
	return opts
}

func (sx *server) String() string {
	return fmt.Sprintf("server(iface=%s, ip=%s, lease_opts=%+v)", sx.iface.Name, sx.selfIP, sx.lopts)
}
