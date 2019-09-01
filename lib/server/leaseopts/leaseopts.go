package leaseopts

import (
	"fmt"
	"net"
	"time"

	pb "gitlab.com/adrian_blx/psa-dhcp/lib/server/proto"
)

type LeaseOptions struct {
	IP            net.IP        // Static IP of a lease.
	Domain        string        // Domain to announce.
	Hostname      string        // Hostname to use.
	Netmask       net.IPMask    // Netmask of the network we announce.
	Router        net.IP        // Router to use.
	DNS           []net.IP      // List of DNS suggested to the client.
	NTP           []net.IP      // List of NTP servers suggested to the client.
	LeaseDuration time.Duration // Duration of the announced lease.
}

// ParseConfig inspects a proto.ServerConfig and returns the configured lease options and the IPNet we are reposible for.
func ParseConfig(conf *pb.ServerConfig) (*LeaseOptions, *net.IPNet, error) {
	_, ipnet, err := net.ParseCIDR(conf.GetNetwork())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse network string '%s': %v", conf.GetNetwork(), err)
	}

	lopts := &LeaseOptions{
		Domain:  conf.GetDomain(),
		Netmask: ipnet.Mask,
	}

	if ld, err := time.ParseDuration(conf.GetLeaseDuration()); err != nil {
		return nil, nil, fmt.Errorf("failed to parse duration from string '%s': %v", conf.GetLeaseDuration(), err)
	} else if ld < time.Minute {
		return nil, nil, fmt.Errorf("lease duration must be at least one minute, found %s", ld)
	} else {
		lopts.LeaseDuration = ld
	}

	if router, err := ipv4(conf.GetRouter()); err != nil {
		return nil, nil, err
	} else if len(router) == 1 {
		lopts.Router = router[0]
	}

	if dns, err := ipv4(conf.GetDns()...); err != nil {
		return nil, nil, err
	} else {
		lopts.DNS = dns
	}

	if ntp, err := ipv4(conf.GetNtp()...); err != nil {
		return nil, nil, err
	} else {
		lopts.NTP = ntp
	}
	return lopts, ipnet, nil
}

// setClientOverrides updates the given leaseOptions pointer and overwrites values with the configuration found in the given clientConfig.
func SetClientOverrides(original *LeaseOptions, client *pb.ClientConfig) error {
	opts := *original
	if ip, err := ipv4(client.GetIp()); err != nil {
		return err
	} else if len(ip) == 1 {
		opts.IP = ip[0]
	}

	if router, err := ipv4(client.GetRouter()); err != nil {
		return err
	} else if len(router) == 1 {
		opts.Router = router[0]
	}

	if dns, err := ipv4(client.GetDns()...); err != nil {
		return err
	} else if len(dns) > 0 {
		opts.DNS = dns
	}

	if ntp, err := ipv4(client.GetNtp()...); err != nil {
		return err
	} else if len(ntp) > 0 {
		opts.NTP = ntp
	}

	if hn := client.GetHostname(); hn != "" {
		opts.Hostname = hn
	}
	// all done, update original reference.
	*original = opts
	return nil
}

func ipv4(list ...string) ([]net.IP, error) {
	if len(list) == 1 && list[0] == "" {
		return nil, nil
	}

	var res []net.IP
	for _, ip := range list {
		if x := net.ParseIP(ip); x != nil && x.To4() != nil {
			res = append(res, x.To4())
		} else {
			return nil, fmt.Errorf("%s is not a valid ipv4", x)
		}
	}
	return res, nil
}
