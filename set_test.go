package goetcd

import (
	"testing"
)

func TestSet(t *testing.T) {
	cluster := "127.0.0.1:4001"
	result, err := Set(cluster, "foo", "bar", 100)

	if err != nil || result.Key != "/foo" || result.Value != "bar" || result.TTL != 99 {
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("Set failed with %s %s %v", result.Key, result.Value, result.TTL)
	}
}
