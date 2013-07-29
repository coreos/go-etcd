package etcd

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/store"
	"io/ioutil"
	"net/url"
	"path"
)

func Set(key string, value string, ttl uint64) (*store.Response, error) {
	logger.Debug("set ", key, ",", value, ",ttl:", ttl, " [", client.cluster.Leader, "]")
	v := url.Values{}
	v.Set("value", value)

	if ttl > 0 {
		v.Set("ttl", fmt.Sprintf("%v", ttl))
	}

	resp, err := sendRequest("POST", path.Join("keys", key), v.Encode())

	if err != nil {
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
