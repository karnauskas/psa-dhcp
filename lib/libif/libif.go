package libif

import (
	"net"
	"time"

	"github.com/vishvananda/netlink"
)

type Ifconfig struct {
	Interface     *net.Interface
	Router        net.IP
	IP            net.IP
	MTU           int
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

	if addrs, err := netlink.AddrList(link, netlink.FAMILY_V4); err != nil {
		return err
	} else {
		for _, addr := range addrs {
			if addr.Label == iface.Name {
				// Only nuke addrs on the main interface, don't touch ethX:N.
				netlink.AddrDel(link, &addr)
			}
		}
	}

	if route, err := defaultRoute(iface); err != nil {
		return err
	} else if route != nil {
		if err := netlink.RouteDel(route); err != nil {
			return err
		}
	}
	return nil
}

// defaultRoute returns the currently configured IPv4 route.
func defaultRoute(iface *net.Interface) (*netlink.Route, error) {
	link, err := setupNL(iface)
	if err != nil {
		return nil, err
	}

	if routes, err := netlink.RouteList(link, netlink.FAMILY_V4); err != nil {
		return nil, err
	} else {
		for _, r := range routes {
			if r.Src == nil && r.Dst == nil && r.Gw != nil {
				return &r, nil
			}
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
