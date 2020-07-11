package libif

import (
	"fmt"
	"net"
	"time"

	"github.com/vishvananda/netlink"
)

type Ifconfig struct {
	Interface     *net.Interface
	Router        net.IP
	IP            net.IP
	MTU           int
	DNS           []net.IP
	DomainName    string
	Netmask       net.IPMask
	LeaseDuration time.Duration
}

// Down attempts to bring an interface down.
func Down(iface *net.Interface) error {
	link, err := setupNL(iface)
	if err != nil {
		return err
	}

	return netlink.LinkSetDown(link)
}

// Up attemtps to bring an interface up.
func Up(iface *net.Interface) error {
	link, err := setupNL(iface)
	if err != nil {
		return err
	}

	return netlink.LinkSetUp(link)
}

// Unconfigure removes the configuration of an interface.
func Unconfigure(iface *net.Interface) error {
	link, err := setupNL(iface)
	if err != nil {
		return err
	}

	addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		if addr.Label == iface.Name {
			// Only nuke addrs on the main interface, don't touch ethX:N.
			netlink.AddrDel(link, &addr)
		}
	}

	route, err := defaultRoute(iface)
	if err != nil {
		return err
	}
	if route != nil {
		err = netlink.RouteDel(route)
		if err != nil {
			return err
		}
	}
	return nil
}

func InterfaceAddr(iface *net.Interface) (net.IP, error) {
	link, err := setupNL(iface)
	if err != nil {
		return nil, err
	}

	addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if addr.Label == iface.Name {
			return addr.IP, nil
		}
	}
	return nil, fmt.Errorf("no ipv4 addr found on interface")
}

// defaultRoute returns the currently configured IPv4 route.
func defaultRoute(iface *net.Interface) (*netlink.Route, error) {
	link, err := setupNL(iface)
	if err != nil {
		return nil, err
	}

	routes, err := netlink.RouteList(link, netlink.FAMILY_V4)
	if err != nil {
		return nil, err
	}

	for _, r := range routes {
		if r.Src == nil && r.Dst == nil && r.Gw != nil {
			return &r, nil
		}
	}
	return nil, nil
}

// SetIface applies an interface configuration.
func SetIface(c Ifconfig) error {
	link, err := setupNL(c.Interface)
	if err != nil {
		return err
	}

	addr := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   c.IP,
			Mask: c.Netmask,
		},
		PreferedLft: int(c.LeaseDuration.Seconds()),
		ValidLft:    int(c.LeaseDuration.Seconds()),
	}
	if err := netlink.AddrReplace(link, addr); err != nil {
		return err
	}

	oldRoute, err := defaultRoute(c.Interface)
	if err != nil {
		return err
	}

	if c.Router != nil {
		if oldRoute != nil && !oldRoute.Gw.Equal(c.Router) {
			// We have an *existing* route which doesn't match: delete it.
			if err := netlink.RouteDel(oldRoute); err != nil {
				return err
			}
		}
		if oldRoute == nil || !oldRoute.Gw.Equal(c.Router) {
			// Non-existing or old route is wrong: add new route.
			newRoute := &netlink.Route{
				LinkIndex: c.Interface.Index,
				Gw:        c.Router,
			}
			if err := netlink.RouteAdd(newRoute); err != nil {
				return err
			}
		}
	}

	if c.MTU > 0 {
		if err := netlink.LinkSetMTU(link, c.MTU); err != nil {
			return err
		}
	}

	return nil
}

func setupNL(iface *net.Interface) (netlink.Link, error) {
	return netlink.LinkByIndex(iface.Index)
}
