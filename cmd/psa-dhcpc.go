package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"

	"gitlab.com/adrian_blx/psa-dhcp/lib/client"
)

var (
	ifname = flag.String("iface", "wlp2s0", "Interface to use")
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l := log.New(os.Stdout, "psa-dhcpc: ", 0)

	iface, err := net.InterfaceByName(*ifname)
	if err != nil {
		l.Fatalf("failed to discover interface %s: %v\n", *iface, err)
	}

	err = client.Run(ctx, l, iface)
	if err != nil {
		l.Fatalf("error: %v\n", err)
	}
}
