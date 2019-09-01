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
		router:  net.ParseIP(conf.GetRouter()),
	}

	if ld, err := time.ParseDuration(conf.GetLeaseDuration()); err != nil {
		return nil, nil, fmt.Errorf("failed to parse duration from string '%s': %v", conf.GetLeaseDuration(), err)
	} else if ld < time.Minute {
		return nil, nil, fmt.Errorf("lease duration must be at least one minute, found %s", ld)
	} else {
		lopts.leaseDuration = ld
	}

	for _, x := range conf.GetDns() {
		if v := net.ParseIP(x); v == nil || v.To4() == nil {
			return nil, nil, fmt.Errorf("failed to parse dns entry '%s' as IPv4", x)
		} else {
			lopts.dns = append(lopts.dns, v)
		}
	}
	for _, x := range conf.GetNtp() {
		if v := net.ParseIP(x); v == nil || v.To4() == nil {
			return nil, nil, fmt.Errorf("failed to parse ntp entry '%s' as IPv4", x)
		} else {
			lopts.ntp = append(lopts.ntp, v)
		}
	}

	// FIXME: do dynamic_range
	// FIXME: do static_only
	// FIXME: do overrides
	return lopts, ipnet, nil
}
