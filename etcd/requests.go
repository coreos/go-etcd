package etcd

import (
	"fmt"
	"net/url"
	"path"
	"reflect"
)

// Valid options for GET, PUT, POST, DELETE
var (
	VALID_GET_OPTIONS = validOptions{
		"recursive":  reflect.Bool,
		"consistent": reflect.Bool,
		"sorted":     reflect.Bool,
		"wait":       reflect.Bool,
		"waitIndex":  reflect.Uint64,
	}

	VALID_PUT_OPTIONS = validOptions{
		"prevValue": reflect.String,
		"prevIndex": reflect.Uint64,
		"prevExist": reflect.Bool,
	}

	VALID_POST_OPTIONS = validOptions{}

	VALID_DELETE_OPTIONS = validOptions{
		"recursive": reflect.Bool,
	}
)

// get issues a GET request
func (c *Client) get(key string, options options) (*Response, error) {
	logger.Debugf("get %s [%s]", key, c.cluster.Leader)

	p := path.Join("keys", key)
	if options != nil {
		str, err := optionsToString(options, VALID_GET_OPTIONS)
		if err != nil {
			return nil, err
		}
		p += str
	}

	resp, err := c.sendRequest("GET", p, "")

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// put issues a PUT request
func (c *Client) put(key string, value string, ttl uint64, options options) (*Response, error) {
	logger.Debugf("put %s, %s, ttl: %d, [%s]", key, value, ttl, c.cluster.Leader)
	v := url.Values{}

	if value != "" {
		v.Set("value", value)
	}

	if ttl > 0 {
		v.Set("ttl", fmt.Sprintf("%v", ttl))
	}

	p := path.Join("keys", key)
	if options != nil {
		str, err := optionsToString(options, VALID_PUT_OPTIONS)
		if err != nil {
			return nil, err
		}
		p += str
	}

	resp, err := c.sendRequest("PUT", p, v.Encode())

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// post issues a POST request
func (c *Client) post(key string, value string, ttl uint64) (*Response, error) {
	logger.Debugf("post %s, %s, ttl: %d, [%s]", key, value, ttl, c.cluster.Leader)
	v := url.Values{}

	if value != "" {
		v.Set("value", value)
	}

	if ttl > 0 {
		v.Set("ttl", fmt.Sprintf("%v", ttl))
	}

	resp, err := c.sendRequest("POST", path.Join("keys", key), v.Encode())

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// delete issues a DELETE request
func (c *Client) delete(key string, options options) (*Response, error) {
	logger.Debugf("delete %s [%s]", key, c.cluster.Leader)
	v := url.Values{}

	p := path.Join("keys", key)
	if options != nil {
		str, err := optionsToString(options, VALID_DELETE_OPTIONS)
		if err != nil {
			return nil, err
		}
		p += str
	}

	resp, err := c.sendRequest("DELETE", p, v.Encode())

	if err != nil {
		return nil, err
	}

	return resp, nil
}
