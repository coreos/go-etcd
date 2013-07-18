package etcd

import (
	"fmt"
	"github.com/coreos/etcd/store"
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	go setHelper("bar")

	result, err := Watch("watch_foo", 0, nil, nil)

	if err != nil || result.Key != "/watch_foo/foo" || result.Value != "bar" {
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("Watch failed with %s %s %v %v", result.Key, result.Value, result.TTL, result.Index)
	}

	result, err = Watch("watch_foo", result.Index, nil, nil)

	if err != nil || result.Key != "/watch_foo/foo" || result.Value != "bar" {
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("Watch with Index failed with %s %s %v %v", result.Key, result.Value, result.TTL, result.Index)
	}

	c := make(chan *store.Response, 10)
	stop := make(chan bool, 1)

	go setLoop("bar")

	go reciver(c, &stop)

	Watch("watch_foo", 0, c, &stop)
}

func setHelper(value string) {
	time.Sleep(time.Second)
	Set("watch_foo/foo", value, 100)
}

func setLoop(value string) {
	time.Sleep(time.Second)
	for i := 0; i < 10; i++ {
		newValue := fmt.Sprintf("%s_%v", value, i)
		Set("watch_foo/foo", newValue, 100)
		time.Sleep(time.Second / 10)
	}
}

func reciver(c chan *store.Response, stop *chan bool) {
	for i := 0; i < 10; i++ {
		fmt.Println(i)
		<-c
	}
	(*stop) <- true
}
