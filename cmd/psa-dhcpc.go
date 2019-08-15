package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/client"
)

var (
	ifname = flag.String("iface", "wlp2s0", "Interface to use")
)

func main() {
	flag.Parse()

	rand.Seed(time.Now().Unix())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l := log.New(os.Stdout, "psa-dhcpc: ", 0)

	iface, err := net.InterfaceByName(*ifname)
	if err != nil {
		l.Fatalf("failed to discover interface %s: %v\n", *ifname, err)
	}

	c := client.New(ctx, l, iface)
	err = c.Run()
	if err != nil {
		l.Fatalf("error: %v\n", err)
	}
}
