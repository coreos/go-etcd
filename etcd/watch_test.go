package etcd

import (
	"fmt"
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	c := NewClient(nil)
	defer func() {
		c.Delete("watch_foo", true)
	}()

	go setHelper("watch_foo", "bar", c)

	resp, err := c.Watch("watch_foo", 0, false, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !(resp.Node.Key == "/watch_foo" && resp.Node.Value == "bar") {
		t.Fatalf("Watch 1 failed: %#v", resp)
	}

	go setHelper("watch_foo", "bar", c)

	resp, err = c.Watch("watch_foo", resp.Node.ModifiedIndex+1, false, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !(resp.Node.Key == "/watch_foo" && resp.Node.Value == "bar") {
		t.Fatalf("Watch 2 failed: %#v", resp)
	}

	ch := make(chan *Response, 10)
	stop := make(chan bool, 1)

	go setLoop("watch_foo", "bar", c)

	go receiver(ch, stop)

	_, err = c.Watch("watch_foo", 0, false, ch, stop)
	if err != ErrWatchStoppedByUser {
		t.Fatalf("Watch returned a non-user stop error")
	}
}

func TestAsyncWatch(t *testing.T) {
	c := NewClient(nil)

	responseChan, _, errorsChan, err := c.AsyncWatch("watch_baz", 0, true)
	if err != nil {
		t.Fatal(err)
	}

	setHelper("watch_baz/foo", "bar", c)

	select {
	case resp := <-responseChan:
		if !(resp.Node.Key == "/watch_baz/foo" && resp.Node.Value == "bar") {
			t.Fatalf("Watch failed")
		}
	case err := <-errorsChan:
		t.Fatal(err)
	case <-time.After(time.Second):
		t.Fatalf("Timed out waiting for a watch response")
	}

	responseChan, stop, errorsChan, err := c.AsyncWatch("watch_baz", 0, true)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Millisecond) //stop only works once the request is in-flight...

	select {
	case stop <- true:
	case <-time.After(time.Second):
		t.Fatalf("Failed to stop....")
	}

	setHelper("watch_baz/wibble", "blah", c)

	select {
	case _, open := <-responseChan:
		if open {
			t.Fatalf("Expected responseChan to be closed")
		}
	case <-time.After(time.Second):
		t.Fatalf("Expected responseChan to close")
	}

	select {
	case _, open := <-errorsChan:
		if open {
			t.Fatalf("Expected errorsChan to be closed")
		}
	case <-time.After(time.Second):
		t.Fatalf("Expected errorsChan to close")
	}
}

func TestWatchAll(t *testing.T) {
	c := NewClient(nil)
	defer func() {
		c.Delete("watch_foo", true)
	}()

	go setHelper("watch_foo/foo", "bar", c)

	resp, err := c.Watch("watch_foo", 0, true, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !(resp.Node.Key == "/watch_foo/foo" && resp.Node.Value == "bar") {
		t.Fatalf("WatchAll 1 failed: %#v", resp)
	}

	go setHelper("watch_foo/foo", "bar", c)

	resp, err = c.Watch("watch_foo", resp.Node.ModifiedIndex+1, true, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !(resp.Node.Key == "/watch_foo/foo" && resp.Node.Value == "bar") {
		t.Fatalf("WatchAll 2 failed: %#v", resp)
	}

	ch := make(chan *Response, 10)
	stop := make(chan bool, 1)

	go setLoop("watch_foo/foo", "bar", c)

	go receiver(ch, stop)

	_, err = c.Watch("watch_foo", 0, true, ch, stop)
	if err != ErrWatchStoppedByUser {
		t.Fatalf("Watch returned a non-user stop error")
	}
}

func setHelper(key, value string, c *Client) {
	time.Sleep(time.Second)
	c.Set(key, value, 100)
}

func setLoop(key, value string, c *Client) {
	time.Sleep(time.Second)
	for i := 0; i < 10; i++ {
		newValue := fmt.Sprintf("%s_%v", value, i)
		c.Set(key, newValue, 100)
		time.Sleep(time.Second / 10)
	}
}

func receiver(c chan *Response, stop chan bool) {
	for i := 0; i < 10; i++ {
		<-c
	}
	stop <- true
}
