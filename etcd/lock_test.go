package etcd

import (
	"testing"
)

type lockResponse struct {
	lockId string
	err    error
}

// Test Lock API
func TestLock(t *testing.T) {

	lockNode := "boo"
	c := NewClient(nil)

	lockId, err := c.Lock(lockNode, "", 1)
	if err != nil {
		t.Fatalf("Could not create lock object foo: %s", err)
	}

	err = c.LockRenew(lockNode, "", 30, lockId)
	if err != nil {
		t.Fatalf("Could not renew lock object foo: %s", err)
	}

	// check that locking by another client isn't possible
	lock2Aquired := make(chan lockResponse)
	var response lockResponse
	go func() {
		c := NewClient(nil)
		lockId2, err := c.Lock(lockNode, "", 1)
		lock2Aquired <- lockResponse{lockId: lockId2, err: err}
	}()

	select {
	case response = <-lock2Aquired:
		t.Fatalf("Lock 2 was aquired when it should have waited for lock 1 to release")
	default:
	}
	err = c.LockDelete(lockNode, "", lockId)
	if err != nil {
		t.Fatalf("Could not remove lock: %s", err)
	}
	err = c.LockDelete(lockNode, "", response.lockId)
	if err != nil {
		t.Fatalf("Could not remove lock: %s", err)
	}
}
