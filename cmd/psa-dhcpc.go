package main

import (
	"context"
	cr "crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"git.sr.ht/~adrian-blx/psa-dhcp/lib/client"
	"git.sr.ht/~adrian-blx/psa-dhcp/lib/resolvconf"
)

var (
	ifname  = flag.String("ifname", "", "Interface to use")
	logTime = flag.Bool("log_time", true, "Prefix log messages with timestamp")
	script  = flag.String("script", "", "Script to execute on significant changes")
	route   = flag.Bool("default_route", true, "Configure (default) route")
	syshook = flag.Bool("syshook", false, "For use in -script: update /etc/resolv.conf")
)

func init() {
	seed := time.Now().UnixNano()
	buf := make([]byte, 8)
	if _, err := cr.Read(buf); err == nil {
		seed += int64(binary.LittleEndian.Uint64(buf))
	}
	rand.Seed(seed)
}

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var lflags int
	if *logTime {
		lflags |= log.LstdFlags
	}
	l := log.New(os.Stdout, "psa-dhcpc: ", lflags)

	if *syshook {
		if err := resolvconf.Run(ctx, l); err != nil {
			l.Fatalf("error running resolvconf hook: %v", err)
		}
		os.Exit(0)
	}

	if *ifname == "" {
		l.Fatalf("-ifname must be set")
	}

	iface, err := net.InterfaceByName(*ifname)
	if err != nil {
		l.Fatalf("failed to discover interface %s: %v\n", *ifname, err)
	}

	l.SetPrefix(fmt.Sprintf("psa-dhcpc[%s] ", iface.Name))

	c := client.New(ctx, l, iface, *script, *route)
	err = c.Run()
	if err != nil {
		l.Fatalf("error: %v\n", err)
	}
}
