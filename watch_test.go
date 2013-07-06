package goetcd

import (
	"fmt"
	"github.com/xiangli-cmu/raft-etcd/store"
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	cluster := "127.0.0.1:4001"

	go setHelper(cluster, "bar")

	result, err := Watch(cluster, "foo", 0, nil)

	if err != nil || result.Key != "/foo/foo" || result.Value != "bar" {
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("Watch failed with %s %s %v %v", result.Key, result.Value, result.TTL, result.Index)
	}
	fmt.Println(result.Index)

	result, err = Watch(cluster, "foo", result.Index, nil)

	if err != nil || result.Key != "/foo/foo" || result.Value != "bar" {
		if err != nil {
			t.Fatal(err)
		}
		t.Fatalf("Watch with Index failed with %s %s %v %v", result.Key, result.Value, result.TTL, result.Index)
	}

	c := make(chan *store.Response, 10)

	go setLoop(cluster, "bar")
	go reciver(c)

	Watch(cluster, "foo", 0, c)
}

func setHelper(cluster string, value string) {
	time.Sleep(time.Second)
	Set(cluster, "foo/foo", value, 100)
}

func setLoop(cluster string, value string) {
	time.Sleep(time.Second)
	for i := 0; i < 10; i++ {
		newValue := fmt.Sprintf("%s_%v", value, i)
		Set(cluster, "foo/foo", newValue, 100)
		time.Sleep(time.Second / 10)
	}
}

func reciver(c chan *store.Response) {
	for i := 0; i < 10; i++ {
		resp := <-c
		fmt.Println(resp.Index)
	}
}
