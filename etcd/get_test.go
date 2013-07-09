package etcd

import (
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	cluster := "127.0.0.1:4001"

	Set(cluster, "foo", "bar", 100)

	// wait for commit
	time.Sleep(100 * time.Millisecond)

	result, err := Get(cluster, "foo")

	if err != nil || result.Key != "/foo" || result.PrevValue != "bar" {
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("Get failed with %s %s %v", result.Key, result.Value, result.TTL)
	}
}
