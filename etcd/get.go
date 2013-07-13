package etcd

import (
	"encoding/json"
	"errors"
	"github.com/coreos/etcd/store"
	"io/ioutil"
	"net/http"
)

func Get(key string) (*[]store.Response, error) {

	var resp *http.Response
	var err error

	httpPath := getHttpPath("keys", key)

	// if we connect to a follower, we will retry until we found a leader
	for {
		resp, err = client.httpClient.Get(httpPath)

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

	var results []store.Response
	var result store.Response

	err = json.Unmarshal(b, &result)

	if err != nil {
		err = json.Unmarshal(b, &results)

		if err != nil {
			resp.Body.Close()
			return nil, err
		}

	} else {
		results = make([]store.Response, 1)
		results[0] = result
	}

	resp.Body.Close()

	return &results, nil

}
