package goetcd

import (
	"testing"
)

func TestDelete(t *testing.T) {
	cluster := "127.0.0.1:4001"
	Set(cluster, "foo", "bar", 100)
	result, err := Delete(cluster, "foo")
	if err != nil {
		t.Fatal(err)
	}

	if result.PrevValue != "bar" || result.Value != "" || result.Exist != true {
		t.Fatalf("Delete failed with %s %s %v", result.PrevValue,
			result.Value, result.Exist)
	}

}
