package server

import (
	"fmt"
	"net"
	"time"

	pb "gitlab.com/adrian_blx/psa-dhcp/lib/server/proto"
)

// parseConfig inspects a proto.ServerConfig and returns the configured lease options and the IPNet we are reposible for.
func parseConfig(conf *pb.ServerConfig) (*leaseOptions, *net.IPNet, error) {
	_, ipnet, err := net.ParseCIDR(conf.GetNetwork())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse network string '%s': %v", conf.GetNetwork(), err)
	}

	lopts := &leaseOptions{
		domain:  conf.GetDomain(),
		netmask: ipnet.Mask,
	}

	if ld, err := time.ParseDuration(conf.GetLeaseDuration()); err != nil {
		return nil, nil, fmt.Errorf("failed to parse duration from string '%s': %v", conf.GetLeaseDuration(), err)
	} else if ld < time.Minute {
		return nil, nil, fmt.Errorf("lease duration must be at least one minute, found %s", ld)
	} else {
		lopts.leaseDuration = ld
	}

	if router, err := ipv4(conf.GetRouter()); err != nil {
		return nil, nil, err
	} else if len(router) == 1 {
		lopts.router = router[0]
	}

	if dns, err := ipv4(conf.GetDns()...); err != nil {
		return nil, nil, err
	} else {
		lopts.dns = dns
	}

	if ntp, err := ipv4(conf.GetNtp()...); err != nil {
		return nil, nil, err
	} else {
		lopts.ntp = ntp
	}
	return lopts, ipnet, nil
}

// setClientOverrides updates the given leaseOptions pointer and overwrites values with the configuration found in the given clientConfig.
func setClientOverrides(opts *leaseOptions, client *pb.ClientConfig) error {
	if ip, err := ipv4(client.GetIp()); err != nil {
		return err
	} else if len(ip) == 1 {
		opts.ip = ip[0]
	}

	if router, err := ipv4(client.GetRouter()); err != nil {
		return err
	} else if len(router) == 1 {
		opts.router = router[0]
	}

	if dns, err := ipv4(client.GetDns()...); err != nil {
		return err
	} else {
		opts.dns = dns
	}

	if ntp, err := ipv4(client.GetNtp()...); err != nil {
		return err
	} else {
		opts.ntp = ntp
	}

	opts.hostname = client.GetHostname()
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
