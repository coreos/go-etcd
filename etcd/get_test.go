package etcd

import (
	"testing"
	"time"
)

func TestGet(t *testing.T) {

	c := NewClient(nil)

	c.Set("foo", "bar", 100)

	// wait for commit
	time.Sleep(100 * time.Millisecond)

	results, err := c.Get("foo", nil)

	if err != nil {
		t.Fatal(err)
	}

	if results[0].Key != "/foo" || results[0].Value != "bar" {
		t.Fatalf("Get failed with %s %s %v", results[0].Key, results[0].Value, results[0].TTL)
	}

	_, err = c.Get("foo", Options{
		"recursive":  true,
		"wait_index": 1,
	})

	if err != nil {
		t.Fatal(err)
	}

	results, err = c.Get("goo", nil)

	if err == nil {
		t.Fatalf("should not be able to get non-exist key")
	}

	results, err = c.GetFrom("foo", "0.0.0.0:4001", nil)

	if err != nil {
		t.Fatal(err)
	}

	if results[0].Key != "/foo" || results[0].Value != "bar" {
		t.Fatalf("Get failed with %s %s %v", results[0].Key, results[0].Value, results[0].TTL)
	}

	results, err = c.GetFrom("foo", "0.0.0.0:4009", nil)

	if err == nil {
		t.Fatal("should not get from port 4009")
	}

	time.Sleep(1 * time.Second)
}
