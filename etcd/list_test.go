package etcd

import (
	"fmt"
	"testing"
	"time"
)

func TestList(t *testing.T) {
	Set("foo_list/foo", "bar", 100)
	Set("foo_list/fooo", "barbar", 100)
	Set("foo_list/foooo/foo", "barbarbar", 100)

	fmt.Println("list set finish")
	// wait for commit
	time.Sleep(time.Second)

	_, err := Get("foo_list")

	if err != nil {
		t.Fatal(err)
	}

}
