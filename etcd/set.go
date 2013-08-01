package etcd

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/store"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
)

func Set(key string, value string, ttl uint64) (*store.Response, error) {
	logger.Debugf("set %s, %s, ttl: %d, [%s]", key, value, ttl, client.cluster.Leader)
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

	if resp.StatusCode != http.StatusOK {

		return nil, handleError(b)
	}

	return convertSetResponse(b)

}

// SetTo sets the value of the key to a given machine address.
// If the given machine is not available or is not leader it returns an error
// Mainly use for testing purpose.
func SetTo(key string, value string, ttl uint64, addr string) (*store.Response, error) {
	v := url.Values{}
	v.Set("value", value)

	if ttl > 0 {
		v.Set("ttl", fmt.Sprintf("%v", ttl))
	}

	httpPath := createHttpPath(addr, path.Join(version, "keys", key))

	resp, err := client.httpClient.PostForm(httpPath, v)

	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)

	resp.Body.Close()

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, handleError(b)
	}

	return convertSetResponse(b)
}

// convert byte stream to response
func convertSetResponse(b []byte) (*store.Response, error) {
	var result store.Response

	err := json.Unmarshal(b, &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}
