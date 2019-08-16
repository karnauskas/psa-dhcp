package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/client"
)

var (
	ifname  = flag.String("iface", "wlp2s0", "Interface to use")
	logTime = flag.Bool("log_time", false, "Prefix log messages with timestamp")
)

func main() {
	flag.Parse()

	rand.Seed(time.Now().Unix())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var lflags int
	if *logTime {
		lflags |= log.LstdFlags
	}
	l := log.New(os.Stdout, "psa-dhcpc: ", lflags)

	iface, err := net.InterfaceByName(*ifname)
	if err != nil {
		l.Fatalf("failed to discover interface %s: %v\n", *ifname, err)
	}

	l.SetPrefix(fmt.Sprintf("psa-dhcpc[%s] ", iface.Name))

	c := client.New(ctx, l, iface)
	err = c.Run()
	if err != nil {
		l.Fatalf("error: %v\n", err)
	}
}
