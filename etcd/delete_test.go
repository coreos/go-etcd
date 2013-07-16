package etcd

import (
	"testing"
)

func TestDelete(t *testing.T) {
	Set("foo", "bar", 100)
	result, err := Delete("foo")
	if err != nil {
		t.Fatal(err)
	}

	if result.PrevValue != "bar" || result.Value != "" {
		t.Fatalf("Delete failed with %s %s", result.PrevValue,
			result.Value)
	}

}
