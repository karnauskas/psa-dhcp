package ipdb

import (
	"net"
	"testing"
	"time"
)

var (
	now        = time.Unix(1000, 0)
	then       = time.Unix(5000, 0)
	leaseLong  = time.Unix(9999, 0)
	leaseShort = time.Unix(1100, 0)
)

func TestClients(t *testing.T) {
	now := time.Unix(1000, 0)
	c := make(clients)

	// First ever client: inject expected to work.
	if err := c.Inject(now, client{ip: uip(1), hwaddr: net.HardwareAddr{0x01}, leasedUntil: leaseLong}); err != nil {
		t.Errorf("Inject #1 failed with %v; wanted nil.", err)
	}
	// Same client, must not work.
	if err := c.Inject(now, client{ip: uip(1), hwaddr: net.HardwareAddr{0x01}, leasedUntil: leaseLong}); err == nil {
		t.Errorf("Inject #2 failed with nil; wanted non-nil.")
	}
	// Same client uip, must not work.
	if err := c.Inject(now, client{ip: uip(1), hwaddr: net.HardwareAddr{0x02}, leasedUntil: leaseLong}); err == nil {
		t.Errorf("Inject #3 failed with nil; wanted non-nil.")
	}
	// Same hwaddr, must not work.
	if err := c.Inject(now, client{ip: uip(2), hwaddr: net.HardwareAddr{0x01}, leasedUntil: leaseLong}); err == nil {
		t.Errorf("Inject #4 failed with nil; wanted non-nil.")
	}
	// Completely new client, inject should work.
	if err := c.Inject(now, client{ip: uip(9), hwaddr: net.HardwareAddr{0x99}, leasedUntil: leaseShort}); err != nil {
		t.Errorf("Inject #5 failed with %v; wanted nil.", err)
	}

	// This client was never injected.
	if a, b := c.Lookup(now, uip(10), net.HardwareAddr{0x10}); a != nil || b != nil {
		t.Errorf("Looking up non existing client worked? a=%v, b=%v\n", a, b)
	}
	// Both are present, but should not return the same value.
	if a, b := c.Lookup(now, uip(9), net.HardwareAddr{0x1}); a == nil || b == nil || a == b {
		t.Errorf("Looking expected to find two different clients, got: a=%p, b=%p\n", a, b)
	}
	// Try again at a later timestamp, entry for IP should be expired.
	if a, b := c.Lookup(then, uip(9), net.HardwareAddr{0x1}); a != nil || b == nil {
		t.Errorf("Looking expected only 'a' to be expired, got a=%p, b=%p\n", a, b)
	}
	// Jumping back to 'now' should now have the same result (we expect the map to be cleaned up)
	if a, b := c.Lookup(now, uip(9), net.HardwareAddr{0x1}); a != nil || b == nil {
		t.Errorf("Looking expected map to be purged, got a=%p, b=%p\n", a, b)
	}

}
