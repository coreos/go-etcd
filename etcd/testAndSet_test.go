package goetcd

import (
	"fmt"
	"testing"
	"time"
)

func TestTestAndSet(t *testing.T) {
	cluster := "127.0.0.1:4001"
	Set(cluster, "foo_testAndSet", "bar", 100)

	time.Sleep(time.Second)

	results := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go testAndSet(cluster, "foo_testAndSet", "bar", "barbar", results)
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

func testAndSet(cluster string, key string, prevValue string, value string, c chan bool) {
	_, success, err := TestAndSet(cluster, key, prevValue, value, 0)
	if err != nil {
		fmt.Println(err)
	}
	c <- success
}
