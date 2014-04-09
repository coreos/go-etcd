package etcd

import (
	"testing"
	"time"
)

type leaderLockResponse struct {
	key  string
	name string
	err  error
}

// Test Leader API
func TestLeader(t *testing.T) {

	c := NewClient(nil)

	process := "orderProcessing"
	value := "node0"

	leaderKey, err := c.LeaderLock(process, value, 1)
	if err != nil {
		t.Fatal("unexpected errors locking leader %s:%s, err:%s", process, value, err)
	}

	t.Logf("Got leader key: %s", leaderKey)

	leaderVal, err := c.LeaderValue(process)
	if err != nil {
		t.Fatal("unexpected error looking up leader value %s, err:%s", process, err)
	}

	if leaderVal != value {
		t.Fatal("expected value %s, got %s", value, leaderVal)
	}

	value2 := "node1"
	secondLeaderChan := make(chan leaderLockResponse)
	go func() {
		c := NewClient(nil)
		leaderKey, err := c.LeaderLock(process, value2, 1)
		if err != nil {
			t.Fatal("unexpected error locking leader %s:%s, err:%s", process, value, err)
		}
		secondLeaderChan <- leaderLockResponse{key: leaderKey, name: value2, err: err}
	}()

	select {
	case response := <-secondLeaderChan:
		if response.err == nil {
			t.Fatalf("Should not have locked by second node: %v", response)
		}
		t.Fatalf("unexpected error when attempting lock by second node: %s", response.err)
	default:
		t.Log("good, second node blocked")
	}

	err = c.LeaderDelete(process, value)
	if err != nil {
		t.Fatal("unexpected error removing leader %s:%s, err:%s", process, value, err)
	}

	select {
	case response := <-secondLeaderChan:
		if response.err != nil {
			t.Fatalf("unexpected error from second node: %v", response.err)
		}
		err = c.LeaderDelete(process, response.name)
		if err != nil {
			t.Fatal("unexpected error removing leader %s:%s, err:%s", response.key, response.name, err)
		}
	case <-time.After(time.Second * 10):
		t.Fatal("expected second node to aquire lock but it timed out")
	}

}
