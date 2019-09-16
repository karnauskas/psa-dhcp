package callback

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/libif"
)

const (
	scriptTimeout = 1 * time.Minute
)

var (
	reSafeChars = regexp.MustCompile(`[^a-zA-Z0-9\.-]`)
)

// Cbhandler returns a function which can be called to execute the specified script.
func Cbhandler(script string, iface *net.Interface, l *log.Logger) func(context.Context, *libif.Ifconfig) {
	return func(ctx context.Context, c *libif.Ifconfig) {
		cargs, err := parseScriptArgs(script)
		if err != nil {
			l.Printf("failed to parse '%s': %v", script, err)
			return
		}
		if len(cargs) == 0 {
			return
		}

		cctx, ccancel := context.WithTimeout(ctx, scriptTimeout)
		defer ccancel()

		cmd := exec.CommandContext(cctx, cargs[0], cargs[1:]...)
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

// dumpScriptConf returns a (shell safe) version of the specified interface configuration.
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
	val = reSafeChars.ReplaceAllString(val, "_")
	return fmt.Sprintf("PSA_DHCPC_%s=%s", key, val)
}

func parseScriptArgs(cmdline string) ([]string, error) {
	r := csv.NewReader(strings.NewReader(cmdline))
	r.Comma = ' '
	r.TrimLeadingSpace = true
	return r.Read()
}
