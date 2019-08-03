package rsocks

import (
	"net"
	"syscall"
)

type rssock struct {
	fd  int
	sll syscall.Sockaddr
}

func GetRawSendSock(iface *net.Interface) (*rssock, error) {
	s, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, int(htons(syscall.ETH_P_IP)))
	if err != nil {
		return nil, err
	}

	sll := &syscall.SockaddrLinklayer{
		Protocol: htons(syscall.ETH_P_IP),
		Halen:    6,
		Addr:     [8]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x0, 0x0},
		Ifindex:  iface.Index,
	}
	if err := syscall.Bind(s, sll); err != nil {
		syscall.Close(s)
		return nil, err
	}
	return &rssock{fd: s, sll: sll}, nil
}

// Write implements the Writer interface.
func (rs *rssock) Write(p []byte) (n int, err error) {
	return len(p), syscall.Sendto(rs.fd, p, 0, rs.sll)
}

// Close closes the socket.
func (rs *rssock) Close() {
	syscall.Close(rs.fd)
}
