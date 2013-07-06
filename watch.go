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

func Watch(cluster string, key string, sinceIndex uint64, receiver chan *store.Response) (*store.Response, error) {

	if receiver == nil {
		return watchOnce(cluster, key, sinceIndex)

	} else {
		for {
			resp, err := watchOnce(cluster, key, sinceIndex)
			if resp != nil {
				receiver <- resp
			} else {
				return nil, err
			}
		}
	}

	// for complier
	return nil, nil

}

func watchOnce(cluster string, key string, sinceIndex uint64) (*store.Response, error) {

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
			resp, err = http.Post(httpPath, "application/x-www-form-urlencoded", reader)

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
