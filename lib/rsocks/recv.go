package rsocks

import (
	"net"
	"syscall"
)

type rrsock struct {
	fd int
}

func GetRawRecvSock(iface *net.Interface) (*rrsock, error) {
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
	return &rrsock{fd: s}, nil
}

func (rr *rrsock) Read(b []byte) (int, error) {
	return syscall.Read(rr.fd, b)
}

func (rr *rrsock) Close() error {
	return syscall.Close(rr.fd)
}
