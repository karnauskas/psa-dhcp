package callback

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"

	"gitlab.com/adrian_blx/psa-dhcp/lib/libif"
)

func Cbhandler(script string, iface *net.Interface, l *log.Logger) func(*libif.Ifconfig) {
	return func(c *libif.Ifconfig) {
		if script == "" {
			return
		}

		cmd := exec.Command(script)
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("PSA_DHCPC_INTERFACE=%q", iface.Name),
		)
		if c != nil {
			cmd.Env = append(cmd.Env, dumpScriptConf(c)...)
		}

		out, err := cmd.CombinedOutput()
		if err != nil {
			l.Printf("Execution of command '%s' returned error: %v", script, err)
		}
		l.Printf("Output of command (if any): %s", string(out))
	}
}

func dumpScriptConf(c *libif.Ifconfig) []string {
	return append([]string{},
		fmt.Sprintf("PSA_DHCPC_IPV4_ROUTER=%q", c.Router),
		fmt.Sprintf("PSA_DHCPC_IPV4_ADDRESS=%q", c.IP),
		fmt.Sprintf("PSA_DHCPC_MTU=%d", c.MTU),
		fmt.Sprintf("PSA_DHCPC_NETMASK=%q", c.Netmask),
		fmt.Sprintf("PSA_DHCPC_LEASE_SEC=%d", int(c.LeaseDuration.Seconds())),
	)
}
