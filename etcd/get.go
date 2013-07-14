package etcd

import (
	"encoding/json"
	"github.com/coreos/etcd/store"
	"io/ioutil"
)

func Get(key string) (*[]store.Response, error) {

	httpPath := getHttpPath("keys", key)

	resp, err := sendRequest(httpPath, nil, nil)

	if err != nil {
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
