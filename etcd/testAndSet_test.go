package etcd

import (
	"testing"
	"time"
)

func TestTestAndSet(t *testing.T) {
	Set("foo_testAndSet", "bar", 100)

	time.Sleep(time.Second)

	results := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go testAndSet("foo_testAndSet", "bar", "barbar", results)
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

func testAndSet(key string, prevValue string, value string, c chan bool) {
	_, success, _ := TestAndSet(key, prevValue, value, 0)
	c <- success
}
