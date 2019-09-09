package main

import (
	"context"
	cr "crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
	"gitlab.com/adrian_blx/psa-dhcp/lib/server"
	pb "gitlab.com/adrian_blx/psa-dhcp/lib/server/proto"
)

var (
	ifname  = flag.String("ifname", "", "Interface to use")
	config  = flag.String("config", "", "Config file to use")
	logTime = flag.Bool("log_time", true, "Prefix log messages with timestamp")
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
	l := log.New(os.Stdout, "psa-dhcpd: ", lflags)

	if *ifname == "" {
		l.Fatalf("-ifname must be set")
	}
	iface, err := net.InterfaceByName(*ifname)
	if err != nil {
		l.Fatalf("failed to discover interface %s: %v\n", *ifname, err)
	}
	l.SetPrefix(fmt.Sprintf("psa-dhcpd[%s] ", iface.Name))

	confpb, err := loadConfig(*config)
	if err != nil {
		l.Fatalf("%s\n", err)
	}

	s, err := server.New(ctx, l, iface, confpb)
	if err != nil {
		l.Fatalf("failed to create new server: %v\n", err)
	}
	err = s.Run()
	if err != nil {
		l.Fatalf("error: %v\n", err)
	}
}

func loadConfig(path string) (*pb.ServerConfig, error) {
	if path == "" {
		return nil, fmt.Errorf("-config must be set")
	}

	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	data, err := ioutil.ReadAll(fh)
	if err != nil {
		return nil, err
	}

	conf := &pb.ServerConfig{}
	err = proto.UnmarshalText(string(data), conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
