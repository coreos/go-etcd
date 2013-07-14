package etcd

import (
	"encoding/json"
	"github.com/coreos/etcd/store"
	"io/ioutil"
	"net/http"
)

func Delete(key string) (*store.Response, error) {

	httpPath := getHttpPath("keys", key)

	// this cannot fail
	req, _ := http.NewRequest("DELETE", httpPath, nil)

	resp, err := sendRequest(httpPath, req, nil)

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
