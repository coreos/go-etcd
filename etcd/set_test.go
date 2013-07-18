package etcd

import (
	"testing"
	"time"
)

func TestSet(t *testing.T) {
	result, err := Set("foo", "bar", 100)

	if err != nil || result.Key != "/foo" || result.Value != "bar" || result.TTL != 99 {
		if err != nil {
			t.Fatal(err)
		}

		t.Fatalf("Set 1 failed with %s %s %v", result.Key, result.Value, result.TTL)
	}

	time.Sleep(time.Second)

	result, err = Set("foo", "bar", 100)

	if err != nil || result.Key != "/foo" || result.Value != "bar" || result.PrevValue != "bar" || result.TTL != 99 {
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("Set 2 failed with %s %s %v", result.Key, result.Value, result.TTL)
	}
}
