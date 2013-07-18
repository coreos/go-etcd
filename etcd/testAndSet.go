package etcd

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/store"
	"io/ioutil"
	"net/url"
	"path"
)

func TestAndSet(key string, prevValue string, value string, ttl uint64) (*store.Response, bool, error) {

	v := url.Values{}
	v.Set("value", value)
	v.Set("prevValue", prevValue)

	if ttl > 0 {
		v.Set("ttl", fmt.Sprintf("%v", ttl))
	}

	resp, err := sendRequest("POST", path.Join("keys", key), v.Encode())

	if err != nil {
		return nil, false, err
	}

	b, err := ioutil.ReadAll(resp.Body)

	resp.Body.Close()

	if err != nil {

		return nil, false, err
	}

	var result store.Response

	err = json.Unmarshal(b, &result)

	if err != nil {
		return nil, false, err
	}

	if result.PrevValue == prevValue && result.Value == value {

		return &result, true, nil
	}

	return &result, false, nil

}
