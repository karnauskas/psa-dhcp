package layer

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDecodeUDP(t *testing.T) {
	input := []struct {
		data []byte
		fail bool
		want *UDP
	}{
		// empty.
		{
			data: []byte{},
			fail: true,
		},
		// all zero.
		{
			data: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			fail: true,
		},
		// truncated.
		{
			data: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x08, 0x00},
			fail: true,
		},
		// too much data.
		{
			data: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x07, 0x00, 0x00},
			fail: true,
		},
		// header size too long.
		{
			data: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x09, 0x00, 0x00},
			fail: true,
		},
		// no payload.
		{
			data: []byte{0x12, 0x34, 0x56, 0x78, 0x00, 0x08, 0x00, 0x00},
			want: &UDP{SrcPort: 0x1234, DstPort: 0x5678, Data: []byte{}},
		},
		// one byte.
		{
			data: []byte{0x22, 0x33, 0x44, 0x55, 0x00, 0x09, 0x00, 0x00, 0xFE},
			want: &UDP{SrcPort: 0x2233, DstPort: 0x4455, Data: []byte{0xfe}},
		},
	}

	for i, test := range input {
		udp, err := DecodeUDP(test.data)
		if test.fail && err == nil {
			t.Errorf("expected test #%d to fail, but finished", i)
		}
		if !test.fail && err != nil {
			t.Errorf("test #%d unexpectedly failed with: %v", i, err)
		}
		if diff := cmp.Diff(test.want, udp); diff != "" {
			t.Errorf("test #%d had a diff: %s", i, diff)
		}
	}
}

func TestAssembleUDP(t *testing.T) {
	input := []struct {
		want []byte
		udp  *UDP
	}{
		{
			want: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00},
			udp:  &UDP{},
		},
		{
			want: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x0a, 0x00, 0x00, 0xAA, 0xFF},
			udp:  &UDP{Data: []byte{0xAA, 0xFF}},
		},
		{
			want: []byte{0xfa, 0xfe, 0xba, 0xce, 0x00, 0x08, 0x00, 0x00},
			udp:  &UDP{SrcPort: 0xfafe, DstPort: 0xbace},
		},
	}

	for i, test := range input {
		data := test.udp.Assemble()
		if diff := cmp.Diff(test.want, data); diff != "" {
			t.Errorf("test #%d had a diff: %s", i, diff)
		}
	}
}
