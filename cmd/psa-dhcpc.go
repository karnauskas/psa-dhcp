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
	ifname  = flag.String("ifname", "", "Interface to use")
	logTime = flag.Bool("log_time", true, "Prefix log messages with timestamp")
	script  = flag.String("script", "", "Script to execute on significant changes")
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

	if *ifname == "" {
		l.Fatalf("-ifname must be set")
	}

	iface, err := net.InterfaceByName(*ifname)
	if err != nil {
		l.Fatalf("failed to discover interface %s: %v\n", *ifname, err)
	}

	l.SetPrefix(fmt.Sprintf("psa-dhcpc[%s] ", iface.Name))

	c := client.New(ctx, l, iface, *script)
	err = c.Run()
	if err != nil {
		l.Fatalf("error: %v\n", err)
	}
}
