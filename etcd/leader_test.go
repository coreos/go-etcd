package etcd

import (
	"testing"
	"time"
)

type leaderLockResponse struct {
	leader *Leader
	err    error
}

// Test Leader API
func TestLeader(t *testing.T) {

	c := NewClient(nil)

	process := "orderProcessing"
	orderProcessing := c.NewLeader(process, "nodeO")

	err := orderProcessing.Lead(1)
	if err != nil {
		t.Fatal("unexpected errors locking leader %v, err: %s", orderProcessing, err)
	}

	leaderVal, err := orderProcessing.Current()
	if err != nil {
		t.Fatal("unexpected error looking up leader value %v, err:%s", orderProcessing, err)
	}

	if leaderVal != orderProcessing.name {
		t.Fatalf("expected value %s, got %s", orderProcessing.name, leaderVal)
	}

	secondLeaderChan := make(chan leaderLockResponse)
	go func() {
		c := NewClient(nil)
		orderProcessing2 := c.NewLeader(process, "node2")
		err := orderProcessing2.Lead(1)
		if err != nil {
			t.Fatal("unexpected error locking leader %v, err:%s", orderProcessing2, err)
		}
		secondLeaderChan <- leaderLockResponse{leader: orderProcessing2, err: err}
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

	err = orderProcessing.Delete()
	if err != nil {
		t.Fatal("unexpected error removing leader %v, err:%s", orderProcessing, err)
	}

	select {
	case response := <-secondLeaderChan:
		if response.err != nil {
			t.Fatalf("unexpected error from second node: %v", response.err)
		}
		err = response.leader.Delete()
		if err != nil {
			t.Fatal("unexpected error removing leader %v, err:%s", response.leader, err)
		}
	case <-time.After(time.Second * 10):
		t.Fatal("expected second node to aquire lock but it timed out")
	}

}
