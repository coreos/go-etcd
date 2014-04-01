package etcd

import (
	"log"
	"testing"
)

// Proof of concept test
func TestLock(t *testing.T) {

	lockNode := "boo"
	c := NewClient(nil)

	lockId, err := c.Lock(lockNode, "", 1)
	if err != nil {
		t.Fatalf("Could not create lock object foo: %s", err)
	}
	log.Printf("Locked foo: %s", lockId)

	// This should not block
	err = c.LockRenew(lockNode, "", 10, lockId)
	if err != nil {
		t.Fatalf("Could not renew lock object foo: %s", err)
	}

	lockId2, err := c.Lock(lockNode, "", 2)
	if err != nil {
		t.Fatalf("Could not create lock object foo: %s", err)
	}
	log.Printf("Locked foo: %s", lockId2)

	err = c.LockRemove(lockNode, "", lockId2)
	if err != nil {
		t.Fatalf("Could not remove lock: %s", err)
	}
}
