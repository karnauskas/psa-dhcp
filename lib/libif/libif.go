package libif

import (
	"fmt"
	"net"
	"os/exec"
)

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

func SetIface(iface *net.Interface, ip net.IP, gw net.IP, netmask *net.IP) {
	// FIXME: NETMASK.
	xexec("ip", "-4", "addr", "add", ip.String()+"/24", "dev", iface.Name)
	xexec("ip", "-4", "route", "add", "default", "via", gw.String(), "dev", iface.Name)
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
