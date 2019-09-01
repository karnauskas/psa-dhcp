package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/dhcpmsg"
	"gitlab.com/adrian_blx/psa-dhcp/lib/libif"
	"gitlab.com/adrian_blx/psa-dhcp/lib/server/ipdb"
	pb "gitlab.com/adrian_blx/psa-dhcp/lib/server/proto"
)

type server struct {
	ctx    context.Context // Context used by this server.
	l      *log.Logger     // Logger.
	iface  *net.Interface  // Interface we are working on.
	selfIP net.IP          // Our own IP (used as server identifier).
	ipdb   *ipdb.IPDB      // IP database instance.
	lopts  leaseOptions    // Default options for leases.
}

type leaseOptions struct {
	domain        string        // Domain to announce.
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

	// Give ourselfs a permanent fake lease.
	if err := db.AddPermanentClient(selfIP, duidFromHwAddr(iface.HardwareAddr)); err != nil {
		return nil, fmt.Errorf("failed to add own IP (%s) to configured net (%s): %v", selfIP, *ipnet, err)
	}

	return &server{ctx: ctx, l: l, iface: iface, selfIP: selfIP, ipdb: db, lopts: *lopts}, nil
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
