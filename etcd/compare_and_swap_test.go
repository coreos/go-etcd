package etcd

import (
	"testing"
	"time"
)

func TestCompareAndSwap(t *testing.T) {
	c := NewClient(nil)

	c.Set("foo_testAndSet", "bar", 100)

	time.Sleep(time.Second)

	results := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		compareAndSwap("foo_testAndSet", "bar", "barbar", results, c)
	}

	count := 0

	for i := 0; i < 3; i++ {
		result := <-results
		if result {
			count++
		}
	}

	if count != 1 {
		t.Fatalf("test and set fails %v", count)
	}

}

func compareAndSwap(key string, prevValue string, value string, ch chan bool, c *Client) {
	resp, _ := c.CompareAndSwap(key, value, 0, Options{
		"prevValue": prevValue,
	})

	if resp != nil {
		ch <- true
	} else {
		ch <- false
	}
}
