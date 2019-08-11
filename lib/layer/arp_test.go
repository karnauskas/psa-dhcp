package layer

import (
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDecodeARP(t *testing.T) {
	input := []struct {
		data []byte
		fail bool
		want *ARP
	}{
		// empty.
		{
			data: []byte{},
			fail: true,
		}, {
			data: []byte{0x00, 0x01, 0x08, 0x00, 0x06, 0x04, 0x00, 0x01, 0x00, 0x04, 0x20, 0x1f, 0x0d, 0x83, 0xc0, 0xa8,
				0x01, 0x28, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xc0, 0xa8, 0x01, 0x6f},
			want: &ARP{
				Opcode:    ARPOpRequest,
				SenderMAC: []byte{0x00, 0x04, 0x20, 0x1f, 0x0d, 0x83},
				TargetMAC: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SenderIP:  net.IPv4(192, 168, 1, 40),
				TargetIP:  net.IPv4(192, 168, 1, 111),
			},
		},
	}

	for i, test := range input {
		arp, err := DecodeARP(test.data)
		if test.fail && err == nil {
			t.Errorf("expected test #%d to fail, but finished", i)
		}
		if !test.fail && err != nil {
			t.Errorf("test #%d unexpectedly failed with: %v", i, err)
		}
		if diff := cmp.Diff(test.want, arp); diff != "" {
			t.Errorf("test #%d had a diff: %s", i, diff)
		}
	}
}

func TestAssembleARP(t *testing.T) {
	input := []struct {
		want []byte
		arp  *ARP
	}{
		{
			want: []byte{0x00, 0x01, 0x08, 0x00, 0x06, 0x04, 0x00, 0x01, 0x00, 0x04, 0x20, 0x1f, 0x0d, 0x83, 0xc0, 0xa8,
				0x01, 0x28, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xc0, 0xa8, 0x01, 0x6f},
			arp: &ARP{
				Opcode:    ARPOpRequest,
				SenderMAC: []byte{0x00, 0x04, 0x20, 0x1f, 0x0d, 0x83},
				TargetMAC: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SenderIP:  net.IPv4(192, 168, 1, 40),
				TargetIP:  net.IPv4(192, 168, 1, 111),
			},
		},
	}

	for i, test := range input {
		data := test.arp.Assemble()
		if diff := cmp.Diff(test.want, data); diff != "" {
			t.Errorf("test #%d had a diff: %s", i, diff)
		}
	}
}
