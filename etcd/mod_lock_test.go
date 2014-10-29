package etcd

import (
	"testing"
	// "time"
)

func TestModLockAcquireAndRelease(t *testing.T) {
	c := NewClient(nil)
	lockName := "foo"

	index, err := c.ModLock_Acquire("foo", "", 0)
	if err != nil {
		t.Error(err)
	}

	proceed := make(chan bool)
	go func() {
		c.ModLock_Acquire("foo", "", 0)
		proceed <- true
	}()

	select {
	case <-proceed:
		t.Fatalf("Should not be able to acquire the lock at this point")
	default:
		// We don't need to block
	}

	c.ModLock_ReleaseWithIndex(lockName, index)
}
