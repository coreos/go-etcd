package etcd

import (
	"encoding/json"
	"errors"
	"github.com/coreos/etcd/store"
	"io/ioutil"
	"net/http"
	"path"
)

func Get(cluster string, key string) (*store.Response, error) {

	httpPath := path.Join(cluster, "/", version, "/keys/", key)

	//TODO: deal with https
	httpPath = "http://" + httpPath

	var resp *http.Response
	var err error
	// if we connect to a follower, we will retry until we found a leader
	for {
		resp, err = http.Get(httpPath)

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
