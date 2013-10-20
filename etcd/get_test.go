package etcd

import (
	"reflect"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	c := NewClient(nil)

	c.Set("foo", "bar", 5)

	// Wait for commit
	time.Sleep(100 * time.Millisecond)

	result, err := c.Get("foo", false)

	if err != nil {
		t.Fatal(err)
	}

	if result.Key != "/foo" || result.Value != "bar" {
		t.Fatalf("Get failed with %s %s %v", result.Key, result.Value, result.TTL)
	}

	result, err = c.Get("goo", false)
	if err == nil {
		t.Fatalf("should not be able to get non-exist key")
	}
}

func TestGetAll(t *testing.T) {
	c := NewClient(nil)

	c.SetDir("fooDir", 5)
	c.Set("fooDir/k0", "v0", 5)
	c.Set("fooDir/k1", "v1", 5)

	// wait for commit
	time.Sleep(100 * time.Millisecond)

	// Return kv-pairs in sorted order
	result, err := c.Get("fooDir", true)

	if err != nil {
		t.Fatal(err)
	}

	expected := kvPairs{
		KeyValuePair{
			Key:   "/fooDir/k0",
			Value: "v0",
		},
		KeyValuePair{
			Key:   "/fooDir/k1",
			Value: "v1",
		},
	}

	if !reflect.DeepEqual(result.Kvs, expected) {
		t.Fatalf("(actual) %v != (expected) %v", result.Kvs, expected)
	}

	// Test the `recursive` option
	c.SetDir("fooDir/childDir", 5)
	c.Set("fooDir/childDir/k2", "v2", 5)

	// Wait for commit
	time.Sleep(100 * time.Millisecond)

	// Return kv-pairs in sorted order
	result, err = c.GetAll("fooDir", true)

	if err != nil {
		t.Fatal(err)
	}

	expected = kvPairs{
		KeyValuePair{
			Key: "/fooDir/childDir",
			Dir: true,
			KVPairs: kvPairs{
				KeyValuePair{
					Key:   "/fooDir/childDir/k2",
					Value: "v2",
				},
			},
		},
		KeyValuePair{
			Key:   "/fooDir/k0",
			Value: "v0",
		},
		KeyValuePair{
			Key:   "/fooDir/k1",
			Value: "v1",
		},
	}

	if !reflect.DeepEqual(result.Kvs, expected) {
		t.Fatalf("(actual) %v != (expected) %v", result.Kvs)
	}
}
