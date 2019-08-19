package callback

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"

	"gitlab.com/adrian_blx/psa-dhcp/lib/libif"
)

func Cbhandler(script string, iface *net.Interface, l *log.Logger) func(*libif.Ifconfig) {
	return func(c *libif.Ifconfig) {
		if script == "" {
			return
		}

		cmd := exec.Command(script)
		cmd.Env = append(os.Environ(),
			envEntry("INTERFACE", iface.Name),
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
	dns := make([]string, len(c.DNS))
	for i, d := range c.DNS {
		dns[i] = d.String()
	}

	return append([]string{},
		envEntry("IPV4_ROUTER", c.Router.String()),
		envEntry("IPV4_ADDRESS", c.IP.String()),
		envEntry("NETMASK", c.Netmask.String()),
		envEntry("DOMAIN_NAME", c.DomainName),
		envEntry("DNS_LIST", strings.Join(dns, ",")),
		envEntry("MTU", fmt.Sprintf("%d", c.MTU)),
		envEntry("LEASE_SEC", fmt.Sprintf("%d", int(c.LeaseDuration.Seconds()))),
	)
}

func envEntry(key, val string) string {
	return fmt.Sprintf("PSA_DHCPC_%s=%s", key, val)
}
