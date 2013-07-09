package etcd

import (
	"encoding/json"
	"errors"
	"github.com/coreos/etcd/store"
	"io/ioutil"
	"net/http"
	"path"
)

func Delete(cluster string, key string) (*store.Response, error) {
	httpPath := path.Join(cluster, "/", version, "/keys/", key)

	//TODO: deal with https
	httpPath = "http://" + httpPath

	client := &http.Client{}

	var resp *http.Response
	var err error

	for {

		req, err := http.NewRequest("DELETE", httpPath, nil)

		resp, err = client.Do(req)

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

	resp.Body.Close()

	if err != nil {
		return nil, err
	}

	var result store.Response

	err = json.Unmarshal(b, &result)

	if err != nil {
		return nil, err
	}

	return &result, nil

}
