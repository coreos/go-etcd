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

	result, err := c.Get("foo")

	if err != nil {
		t.Fatal(err)
	}

	if result.Key != "/foo" || result.Value != "bar" {
		t.Fatalf("Get failed with %s %s %v", result.Key, result.Value, result.TTL)
	}

	_, err = c.GetDir("foo", false, false)

	if err != nil {
		t.Fatal(err)
	}

	result, err = c.Get("goo")

	if err == nil {
		t.Fatalf("should not be able to get non-exist key")
	}
}
