package goetcd

import (
	"testing"
	"time"
)

func TestList(t *testing.T) {
	cluster := "127.0.0.1:4003"

	Set(cluster, "foo_list/foo", "bar", 100)
	Set(cluster, "foo_list/fooo", "barbar", 100)
	Set(cluster, "foo_list/foooo/foo", "barbarbar", 100)

	// wait for commit
	time.Sleep(time.Second)

	_, err := List(cluster, "foo_list")

	if err != nil {
		t.Fatal(err)
	}

}
