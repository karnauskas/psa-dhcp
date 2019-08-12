package rsocks

import (
	"net"
	"syscall"
)

type rssock struct {
	fd  int
	sll syscall.Sockaddr
}

var (
	bcastAddr = [6]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
)

func GetIPSendSock(iface *net.Interface) (*rssock, error) {
	return getSendSock(iface, htons(syscall.ETH_P_IP), bcastAddr)
}

func GetARPSendSock(iface *net.Interface) (*rssock, error) {
	return getSendSock(iface, htons(syscall.ETH_P_ARP), bcastAddr)
}

func getSendSock(iface *net.Interface, proto uint16, hwaddr [6]byte) (*rssock, error) {
	s, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, int(proto))
	if err != nil {
		return nil, err
	}

	var sllHwaddr [8]byte
	copy(sllHwaddr[0:], hwaddr[0:])

	sll := &syscall.SockaddrLinklayer{
		Protocol: proto,
		Halen:    uint8(len(hwaddr)),
		Addr:     sllHwaddr,
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
func (rs *rssock) Close() error {
	return syscall.Close(rs.fd)
}
