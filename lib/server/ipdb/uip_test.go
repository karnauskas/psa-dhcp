package ipdb

import (
	"net"
	"testing"
)

func TestToV4(t *testing.T) {
	input := []struct {
		got  uip
		want net.IP
	}{
		{got: uip(0xFFFFFFFF), want: net.IPv4(255, 255, 255, 255)},
		{got: uip(0x00000000), want: net.IPv4(0, 0, 0, 0)},
		{got: uip(0xACA80309), want: net.IPv4(172, 168, 3, 9)},
	}

	for i, x := range input {
		if v4 := x.got.toV4(); !v4.Equal(x.want) {
			t.Errorf("test %d produces %+v from %x, wanted %+v", i, v4, x.got, x.want)
		}
	}
}

func TestValid(t *testing.T) {
	input := []struct {
		got   uip
		valid bool
	}{
		{got: uip(0xFFFFFFFF), valid: false},
		{got: uip(0x00000000), valid: false},
		{got: uip(0xACA803FF), valid: false},
		{got: uip(0xACA80309), valid: true},
	}

	for i, x := range input {
		if x.got.valid() != x.valid {
			t.Errorf("test %d %X = %v; want %v", i, x.got, x.got.valid(), x.valid)
		}
	}
}
