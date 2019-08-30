package clients

import (
	"net"
	"testing"
	"time"

	"gitlab.com/adrian_blx/psa-dhcp/lib/server/ipdb/uip"
)

var (
	now        = time.Unix(1000, 0)
	then       = time.Unix(5000, 0)
	leaseLong  = time.Unix(9999, 0)
	leaseShort = time.Unix(1100, 0)
)

func TestClients(t *testing.T) {
	c := NewClients()

	// First ever client: inject expected to work.
	if err := c.Inject(now, uip.Uip(1), net.HardwareAddr{0x01}, leaseLong); err != nil {
		t.Errorf("Inject #1 failed with %v; wanted nil.", err)
	}
	// Same client, must not work.
	if err := c.Inject(now, uip.Uip(1), net.HardwareAddr{0x01}, leaseLong); err == nil {
		t.Errorf("Inject #2 failed with nil; wanted non-nil.")
	}
	// Same client uip, must not work.
	if err := c.Inject(now, uip.Uip(1), net.HardwareAddr{0x02}, leaseLong); err == nil {
		t.Errorf("Inject #3 failed with nil; wanted non-nil.")
	}
	// Same hwaddr, must not work.
	if err := c.Inject(now, uip.Uip(2), net.HardwareAddr{0x01}, leaseLong); err == nil {
		t.Errorf("Inject #4 failed with nil; wanted non-nil.")
	}
	// Completely new client, inject should work.
	if err := c.Inject(now, uip.Uip(9), net.HardwareAddr{0x99}, leaseShort); err != nil {
		t.Errorf("Inject #5 failed with %v; wanted nil.", err)
	}

	// This client was never injected.
	if a, b := c.Lookup(now, uip.Uip(10), net.HardwareAddr{0x10}); a != nil || b != nil {
		t.Errorf("Looking up non existing client worked? a=%v, b=%v\n", a, b)
	}
	// Both are present, but should not return the same value.
	if a, b := c.Lookup(now, uip.Uip(9), net.HardwareAddr{0x1}); a == nil || b == nil || a == b {
		t.Errorf("Looking expected to find two different clients, got: a=%p, b=%p\n", a, b)
	}
	// Try again at a later timestamp, entry for IP should be expired.
	if a, b := c.Lookup(then, uip.Uip(9), net.HardwareAddr{0x1}); a != nil || b == nil {
		t.Errorf("Looking expected only 'a' to be expired, got a=%p, b=%p\n", a, b)
	}
	// Jumping back to 'now' should now have the same result (we expect the map to be cleaned up)
	if a, b := c.Lookup(now, uip.Uip(9), net.HardwareAddr{0x1}); a != nil || b == nil {
		t.Errorf("Looking expected map to be purged, got a=%p, b=%p\n", a, b)
	}
}

func TestPermanent(t *testing.T) {
	c := NewClients()

	// Permanent client with short TTL.
	if err := c.InjectPermanent(now, uip.Uip(9), net.HardwareAddr{0x99}); err != nil {
		t.Errorf("Injecting permanent client failed with %v; wanted nil.", err)
	}
	// Add new client while the permanent one should already be expired; still must fail.
	if err := c.Inject(then, uip.Uip(9), net.HardwareAddr{0x99}, leaseShort); err == nil {
		t.Errorf("Injecting permanent but expired client returned nil, wanted non-nil.")
	}

	// Lookup at 'now' should work.
	if a, b := c.Lookup(now, uip.Uip(9), net.HardwareAddr{0x99}); a == nil || b == nil || a != b {
		t.Errorf("Expected to find permanent client, got a=%p, b=%p", a, b)
	}
	// Lookup at 'then' should also work as permanent entries never expire.
	if a, b := c.Lookup(then, uip.Uip(9), net.HardwareAddr{0x99}); a == nil || b == nil || a != b {
		t.Errorf("Expected to find permanent client, got a=%p, b=%p", a, b)
	}
}

func TestExpireInject(t *testing.T) {
	c := NewClients()

	if err := c.Inject(now, uip.Uip(9), net.HardwareAddr{0x99}, leaseShort); err != nil {
		t.Errorf("Injecting client failed with %v; wanted nil.", err)
	}
	if err := c.Inject(then, uip.Uip(9), net.HardwareAddr{0x99}, leaseShort); err != nil {
		t.Errorf("Injecting 2nd client failed with %v; wanted nil.", err)
	}
}

func TestLeaseSet(t *testing.T) {
	c := NewClients()

	if err := c.Inject(now, uip.Uip(9), net.HardwareAddr{0x99}, leaseShort); err != nil {
		t.Errorf("Injecting client failed with %v; wanted nil.", err)
	}
	if err := c.SetLease(now, uip.Uip(9), net.HardwareAddr{0x99}, leaseLong); err != nil {
		t.Errorf("SetLease failed with %v; wanted nil.", err)
	}
	// Test wrong params
	if err := c.SetLease(now, uip.Uip(9), net.HardwareAddr{0x91}, leaseLong); err == nil {
		t.Errorf("SetLease with invalid hwaddr worked, wanted err")
	}
	if err := c.SetLease(now, uip.Uip(8), net.HardwareAddr{0x99}, leaseLong); err == nil {
		t.Errorf("SetLease with invalid ip worked, wanted err")
	}
	if err := c.SetLease(now, uip.Uip(8), net.HardwareAddr{0x91}, leaseLong); err == nil {
		t.Errorf("SetLease with invalid ip and hwaddr worked, wanted err")
	}
	// Now lookup at later time.
	if a, b := c.Lookup(then, uip.Uip(9), net.HardwareAddr{0x99}); a == nil || a != b {
		t.Errorf("Lookup of extended lease failed")
	}
	// ..and expire it.
	if err := c.Expire(now, uip.Uip(9), net.HardwareAddr{0x99}); err != nil {
		t.Errorf("Expire failed")
	}
	// That should now fail
	if a, b := c.Lookup(then, uip.Uip(9), net.HardwareAddr{0x99}); a != nil || b != nil {
		t.Errorf("Lookup of expire entry did not fail?!")
	}
}
