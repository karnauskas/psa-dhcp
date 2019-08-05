package rsocks

import (
	"net"
	"os"
	"syscall"
)

type rrsock struct {
	fd int
}

func GetRawRecvSock(iface *net.Interface) (*os.File, error) {
	s, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, int(htons(syscall.ETH_P_IP)))
	if err != nil {
		return nil, err
	}
	sll := &syscall.SockaddrLinklayer{
		Protocol: htons(syscall.ETH_P_IP),
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
