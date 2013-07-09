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

var version = "v1"

func Set(cluster string, key string, value string, ttl uint64) (*store.Response, error) {

	httpPath := path.Join(cluster, "/", version, "/keys/", key)

	//TODO: deal with https
	httpPath = "http://" + httpPath

	v := url.Values{}
	v.Set("value", value)

	if ttl > 0 {
		v.Set("ttl", fmt.Sprintf("%v", ttl))
	}

	var resp *http.Response
	var err error
	// if we connect to a follower, we will retry until we found a leader
	for {
		client := http.Client{}
		resp, err = client.PostForm(httpPath, v)

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
