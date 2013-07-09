package etcd

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/etcd/store"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

type respAndErr struct {
	resp *http.Response
	err  error
}

// Watch any change under the given prefix
// When a sinceIndex is given, watch will try to scan from that index to the last index
// and will return any changes under the given prefix during the histroy
// If a receiver channel is given, it will be a long-term watch. Watch will block at the
// channel. And after someone receive the channel, it will go on to watch that prefix.
// If a stop channel is given, client can close long-term watch using the stop channel

func Watch(cluster string, prefix string, sinceIndex uint64, receiver chan *store.Response, stop *chan bool) (*store.Response, error) {

	if receiver == nil {
		return watchOnce(cluster, prefix, sinceIndex, stop)

	} else {
		for {
			resp, err := watchOnce(cluster, prefix, sinceIndex, stop)
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

// helper func
// return when there is change under the given prefix
func watchOnce(cluster string, key string, sinceIndex uint64, stop *chan bool) (*store.Response, error) {

	httpPath := path.Join(cluster, "/", version, "/watch/", key)

	//TODO: deal with https
	httpPath = "http://" + httpPath
	var resp *http.Response
	var err error

	if sinceIndex == 0 {

		// Get request if no index is given
		resp, err = http.Get(httpPath)
		if resp == nil {
			return nil, err
		}

	} else {

		// Post
		v := url.Values{}
		v.Set("index", fmt.Sprintf("%v", sinceIndex))

		// if we connect to a follower, we will retry until we found a leader
		for {

			c := make(chan respAndErr)
			client := http.Client{}
			resp, err = client.PostForm(httpPath, v)

			if stop != nil {
				go func() {
					resp, err = client.PostForm(httpPath, v)

					c <- respAndErr{resp, err}
				}()

				// select at stop or continue to receive
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
