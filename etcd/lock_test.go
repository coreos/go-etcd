package etcd

import (
	"testing"
)

type lockResponse struct {
	lock *Lock
	err  error
}

// Test Lock API
func TestLock(t *testing.T) {

	c := NewClient(nil)
	lockBoo := c.NewLock("boo", "", 10)

	err := lockBoo.Lock()
	if err != nil {
		t.Fatalf("Could not create lock object: %s", err)
	}

	lockBoo.SetTTL(10)

	err = lockBoo.Renew()

	if err != nil {
		t.Fatalf("Could not renew lock object: %s", err)
	}

	// check that locking by another client isn't possible
	lock2Aquired := make(chan lockResponse)
	var response lockResponse

	go func() {
		c := NewClient(nil)
		lockBoo2 := c.NewLock("boo", "", 10)
		err := lockBoo2.Lock()
		lock2Aquired <- lockResponse{lock: lockBoo2, err: err}
	}()

	select {
	case response = <-lock2Aquired:
		t.Fatalf("Lock 2 was aquired when it should have waited for lock 1 to release")
	default:
	}

	err = lockBoo.Delete()
	if err != nil {
		t.Fatalf("Could not remove lock: %s", err)
	}

	response = <-lock2Aquired
	if response.err != nil {
		t.Fatalf("Error aquiring second lock: %s", response.err)
	}
	err = response.lock.Delete()
	if err != nil {
		t.Fatalf("Could not remove lock: %s", err)
	}

}
