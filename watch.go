package goetcd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/xiangli-cmu/raft-etcd/store"
	"io/ioutil"
	"net/http"
	"path"
)

type respAndErr struct {
	resp *http.Response
	err  error
}

func Watch(cluster string, key string, sinceIndex uint64, receiver chan *store.Response, stop *chan bool) (*store.Response, error) {

	if receiver == nil {
		return watchOnce(cluster, key, sinceIndex, stop)

	} else {
		for {
			resp, err := watchOnce(cluster, key, sinceIndex, stop)
			if resp != nil {
				sinceIndex = resp.Index
				receiver <- resp
			} else {
				return nil, err
			}
		}
	}

	// for complier
	return nil, nil

}

func watchOnce(cluster string, key string, sinceIndex uint64, stop *chan bool) (*store.Response, error) {

	httpPath := path.Join(cluster, "/", version, "/watch/", key)

	//TODO: deal with https
	httpPath = "http://" + httpPath
	var resp *http.Response
	var err error

	if sinceIndex == 0 {
		// Get
		resp, err = http.Get(httpPath)
		if resp == nil {
			return nil, err
		}

	} else {
		// Post
		content := fmt.Sprintf("index=%v", sinceIndex)
		// if we connect to a follower, we will retry until we found a leader
		for {
			reader := bytes.NewReader([]byte(content))

			c := make(chan respAndErr)

			if stop == nil {
				resp, err = http.Post(httpPath, "application/x-www-form-urlencoded", reader)
			} else {
				go func() {
					resp, err := http.Post(httpPath, "application/x-www-form-urlencoded", reader)
					c <- respAndErr{resp, err}
				}()
				select {
				case res := <-c:
					resp, err = res.resp, res.err
				case <-(*stop):
					resp, err = nil, errors.New("User stoped watch")
				}
			}

			if resp != nil {

				if resp.StatusCode == http.StatusTemporaryRedirect {
					httpPath = resp.Header.Get("Location")

					resp.Body.Close()

					if httpPath == "" {
						return nil, errors.New("Cannot get redirection location")
					}

					// try to connect the leader
					continue
				} else {
					break
				}

			}

			return nil, err
		}

	}

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		resp.Body.Close()
		return nil, err
	}

	var result store.Response

	err = json.Unmarshal(b, &result)

	if err != nil {
		resp.Body.Close()
		return nil, err
	}
	resp.Body.Close()

	return &result, nil
}
