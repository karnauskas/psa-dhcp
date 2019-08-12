package rsocks

import (
	"net"
	"os"
	"syscall"
)

type rrsock struct {
	fd int
}

func GetIPRecvSock(iface *net.Interface) (*os.File, error) {
	return getRecvSock(iface, htons(syscall.ETH_P_IP))
}

func GetARPRecvSock(iface *net.Interface) (*os.File, error) {
	return getRecvSock(iface, htons(syscall.ETH_P_ARP))
}

func getRecvSock(iface *net.Interface, proto uint16) (*os.File, error) {
	s, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, int(proto))
	if err != nil {
		return nil, err
	}
	sll := &syscall.SockaddrLinklayer{
		Protocol: proto,
		Ifindex:  iface.Index,
	}
	if err := syscall.Bind(s, sll); err != nil {
		syscall.Close(s)
		return nil, err
	}
	if err := syscall.SetNonblock(s, true); err != nil {
		syscall.Close(s)
		return nil, err
	}
	return os.NewFile(uintptr(s), ""), nil
}
