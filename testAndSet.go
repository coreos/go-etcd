package goetcd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/etcd/store"
	"io/ioutil"
	"net/http"
	"path"
)

func TestAndSet(cluster string, key string, prevValue string, value string, ttl uint64) (*store.Response, bool, error) {

	httpPath := path.Join(cluster, "/", version, "/testAndSet/", key)

	//TODO: deal with https
	httpPath = "http://" + httpPath

	content := "value=" + value + "&prevValue=" + prevValue

	if ttl > 0 {
		content += fmt.Sprintf("&ttl=%v", ttl)
	}

	var resp *http.Response
	var err error
	// if we connect to a follower, we will retry until we found a leader
	for {
		reader := bytes.NewReader([]byte(content))
		resp, err = http.Post(httpPath, "application/x-www-form-urlencoded", reader)

		if resp != nil {

			if resp.StatusCode == http.StatusTemporaryRedirect {
				httpPath = resp.Header.Get("Location")

				resp.Body.Close()

				if httpPath == "" {
					return nil, false, errors.New("Cannot get redirection location")
				}

				// try to connect the leader
				continue
			} else {
				break
			}

		}

		return nil, false, err
	}

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		resp.Body.Close()
		return nil, false, err
	}

	var result store.Response

	err = json.Unmarshal(b, &result)

	if err != nil {
		resp.Body.Close()
		return nil, false, err
	}
	resp.Body.Close()

	if result.PrevValue == prevValue && result.Value == value {

		return &result, true, nil
	}

	return &result, false, nil

}
