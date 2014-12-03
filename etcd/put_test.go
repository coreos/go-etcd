package etcd

import (
	"testing"
)

func TestPut(t *testing.T) {
	c := NewClient(nil)
	defer func() {
		c.Delete("foo", true)
	}()

	c.Put("foo", "bar", 5, Options{})

	// This should succeed
	resp, err := c.Put("foo", "bar2", 5, Options{"prevValue": "bar"})
	if err != nil {
		t.Fatal(err)
	}
	if !(resp.Node.Value == "bar2" && resp.Node.Key == "/foo" && resp.Node.TTL == 5) {
		t.Fatalf("CompareAndSwap 1 failed: %#v", resp)
	}

	if !(resp.PrevNode.Value == "bar" && resp.PrevNode.Key == "/foo" && resp.PrevNode.TTL == 5) {
		t.Fatalf("CompareAndSwap 1 prevNode failed: %#v", resp)
	}

	// This should fail because it gives an incorrect prevValue
	resp, err = c.Put("foo", "bar2", 5, Options{"prevValue": "xxx"})
	if err == nil {
		t.Fatalf("CompareAndSwap 2 should have failed.  The response is: %#v", resp)
	}

	resp, err = c.Put("foo", "bar", 5, Options{})
	if err != nil {
		t.Fatal(err)
	}

	// This should succeed
	resp, err = c.Put("foo", "bar2", 5, Options{"prevIndex": resp.Node.ModifiedIndex})
	if err != nil {
		t.Fatal(err)
	}
	if !(resp.Node.Value == "bar2" && resp.Node.Key == "/foo" && resp.Node.TTL == 5) {
		t.Fatalf("CompareAndSwap 3 failed: %#v", resp)
	}

	if !(resp.PrevNode.Value == "bar" && resp.PrevNode.Key == "/foo" && resp.PrevNode.TTL == 5) {
		t.Fatalf("CompareAndSwap 3 prevNode failed: %#v", resp)
	}

	// This should fail because it gives an incorrect prevIndex
	resp, err = c.Put("foo", "bar2", 5, Options{"prevIndex": 29817514})
	if err == nil {
		t.Fatalf("CompareAndSwap 4 should have failed.  The response is: %#v", resp)
	}
}
