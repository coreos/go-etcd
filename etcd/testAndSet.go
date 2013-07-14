package etcd

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/store"
	"io/ioutil"
	"net/url"
)

func TestAndSet(cluster string, key string, prevValue string, value string, ttl uint64) (*store.Response, bool, error) {
	httpPath := getHttpPath("keys", key)

	v := url.Values{}
	v.Set("value", value)
	v.Set("prevValue", prevValue)

	if ttl > 0 {
		v.Set("ttl", fmt.Sprintf("%v", ttl))
	}

	resp, err := sendRequest(httpPath, nil, &v)

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
