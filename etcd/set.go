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

// Create a directory
func (c *Client) SetDir(key string, ttl uint64) (*store.Response, error) {
	return c.Set(key, "", ttl)
}

// Create a key-value pair
func (c *Client) Set(key string, value string, ttl uint64) (*store.Response, error) {
	logger.Debugf("set %s, %s, ttl: %d, [%s]", key, value, ttl, c.cluster.Leader)
	v := url.Values{}

	if value != "" {
		v.Set("value", value)
	}

	if ttl > 0 {
		v.Set("ttl", fmt.Sprintf("%v", ttl))
	}

	resp, err := c.sendRequest("PUT", path.Join("keys", key), v.Encode())

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

// Create a directory on the given machine
func (c *Client) SetDirTo(key string, ttl uint64, addr string) (*store.Response, error) {
	return c.SetTo(key, "", ttl, addr)
}

// SetTo sets the value of the key to a given machine address.
// If the given machine is not available or is not leader it returns an error
// Mainly use for testing purpose.
func (c *Client) SetTo(key string, value string, ttl uint64, addr string) (*store.Response, error) {
	v := url.Values{}

	if value != "" {
		v.Set("value", value)
	}

	if ttl > 0 {
		v.Set("ttl", fmt.Sprintf("%v", ttl))
	}

	httpPath := c.createHttpPath(addr, path.Join(version, "keys", key))
	resp, err := c.sendRequest("PUT", httpPath, v.Encode())

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

// Convert byte stream to response.
func convertSetResponse(b []byte) (*store.Response, error) {
	var result store.Response

	err := json.Unmarshal(b, &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}
