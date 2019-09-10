package oui

import (
	"net"
	"testing"
)

func Test(t *testing.T) {
	input := []struct {
		mac    net.HardwareAddr
		want   string
		exists bool
	}{
		{
			mac:    net.HardwareAddr{0xf4, 0x8c, 0x50, 0xff, 0xff, 0xff},
			want:   "Intel Corporate",
			exists: true,
		},
		{
			mac:    net.HardwareAddr{0xff, 0xff, 0x50, 0xff, 0xff, 0xff},
			want:   "",
			exists: false,
		},
	}

	for _, test := range input {
		got, ok := Lookup(test.mac)
		if got != test.want {
			t.Errorf("Lookup(%v) = %s, want %s", test.mac, got, test.want)
		}
		if ok != test.exists {
			t.Errorf("Lookup(%v) = %v, wanted exists = %v", test.mac, ok, test.exists)
		}
	}
}
