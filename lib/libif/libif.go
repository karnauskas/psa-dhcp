package libif

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"time"
)

type Ifconfig struct {
	Interface     *net.Interface
	Router        net.IP
	IP            net.IP
	Cidr          int
	LeaseDuration time.Duration
}

// Down attempts to bring an interface down.
func Down(iface *net.Interface) error {
	return xexec("ip", "link", "set", "dev", iface.Name, "down")
}

// Up attemtps to bring an interface up.
func Up(iface *net.Interface) error {
	return xexec("ip", "link", "set", "dev", iface.Name, "up")
}

// Unconfigure removes the configuration of an interface.
func Unconfigure(iface *net.Interface) error {
	cmds := [][]string{
		{"ip", "-4", "addr", "del", "dev", iface.Name},
		{"ip", "-4", "route", "del", "default", "dev", iface.Name},
	}

	var lerr error
	for _, cmd := range cmds {
		if err := xexec(cmd...); err != nil {
			lerr = err
		}
	}
	return lerr
}

// DefaultRoute returns the currently configured IPv4 route.
func DefaultRoute(iface *net.Interface) (net.IP, error) {
	c := exec.Command("ip", "-4", "route", "list", "0/0", "dev", iface.Name)
	out, err := c.CombinedOutput()
	if err != nil {
		return net.IP{}, err
	}
	if len(out) == 0 {
		// no route is not really an error.
		return net.IP{}, nil
	}
	re := regexp.MustCompile(`default via (\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	if m := re.FindStringSubmatch(string(out)); len(m) == 2 {
		return net.ParseIP(m[1]), nil
	}
	return net.IP{}, fmt.Errorf("failed to match IP in route output")
}

// SetIface applies an interface configuration.
func SetIface(c Ifconfig) error {
	lft := fmt.Sprintf("%d", int(c.LeaseDuration.Seconds()))

	if err := xexec("ip", "-4", "addr", "replace", fmt.Sprintf("%s/%d", c.IP.String(), c.Cidr),
		"valid_lft", lft, "preferred_lft", lft, "dev", c.Interface.Name); err != nil {
		return err
	}

	oldRoute, err := DefaultRoute(c.Interface)
	if err != nil {
		return err
	}

	if !oldRoute.Equal(net.IP{}) && !oldRoute.Equal(c.Router) {
		// We have an *existing* route which doesn't match: delete it.
		if err := xexec("ip", "-4", "route", "del", "default", "via", oldRoute.String(), "dev", c.Interface.Name); err != nil {
			return err
		}
	}
	if !oldRoute.Equal(c.Router) {
		// Non-existing or old route is wrong: add new route.
		if err := xexec("ip", "-4", "route", "add", "default", "via", c.Router.String(), "dev", c.Interface.Name); err != nil {
			return err
		}
	}
	return nil
}

func xexec(cmd ...string) error {
	if len(cmd) < 2 {
		return fmt.Errorf("short command: %+v", cmd)
	}

	c := exec.Command(cmd[0], cmd[1:]...)
	out, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v, output: %s", err, out)
	}
	return nil
}
