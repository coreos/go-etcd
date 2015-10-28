package etcd

import (
	"testing"
)

func TestCompareAndSwap(t *testing.T) {
	c := NewClient(nil)
	defer func() {
		c.Delete("foo", true)
	}()

	c.Set("foo", "bar", 5)

	// This should succeed
	resp, err := c.CompareAndSwap("foo", "bar2", 5, "bar", 0, Ignored)
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
	resp, err = c.CompareAndSwap("foo", "bar3", 5, "xxx", 0, Ignored)
	if err == nil {
		t.Fatalf("CompareAndSwap 2 should have failed.  The response is: %#v", resp)
	}

	resp, err = c.Set("foo", "bar", 5)
	if err != nil {
		t.Fatal(err)
	}

	// This should succeed
	resp, err = c.CompareAndSwap("foo", "bar2", 5, "", resp.Node.ModifiedIndex, Ignored)
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
	resp, err = c.CompareAndSwap("foo", "bar3", 5, "", 29817514, Ignored)
	if err == nil {
		t.Fatalf("CompareAndSwap 4 should have failed.  The response is: %#v", resp)
	}

    // This should fail because the key already exists
	resp, err = c.CompareAndSwap("foo", "bar4", 5, "", 0, DoesNotExist)
	if err == nil {
		t.Fatalf("CompareAndSwap 5 should have failed. The response is %#v", resp)
	}

	// This should succeed
	c.Set("foo", "bar", 5)
	resp, err = c.CompareAndSwap("foo", "bar5", 5, "", 0, Exists)
	if err != nil {
		t.Fatalf("CompareAndSwap 6 failed: %#v %#v", resp, err)
	}
}
