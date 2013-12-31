package etcd

import (
	"fmt"
	"net/url"
	"strconv"
)

const mod_lock_prefix = "mod/v2/lock/"

func nameToPath(name string) string {
	return mod_lock_prefix + name
}

func toUrlValues(value string, index uint64, ttl uint64) url.Values {
	v := url.Values{}

	if value != "" {
		v.Set("value", value)
	}

	if ttl >= 0 {
		v.Set("ttl", fmt.Sprintf("%v", ttl))
	}

	if index > 0 {
		v.Set("index", fmt.Sprintf("%v", ttl))
	}

	return v
}

func (c *Client) ModLock_Acquire(name string, value string, ttl uint64) (uint64, error) {
	v := toUrlValues(value, 0, ttl)
	raw, err := c.sendRequest("POST", nameToPath(name), v)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(string(raw.Body), 10, 0)
}

func (c *Client) ModLock_RenewWithValue(name string, value string, ttl uint64) error {
	v := toUrlValues(value, 0, ttl)
	_, err := c.sendRequest("PUT", nameToPath(name), v)
	return err
}

func (c *Client) ModLock_RenewWithIndex(name string, index uint64, ttl uint64) error {
	v := toUrlValues("", index, ttl)
	_, err := c.sendRequest("PUT", nameToPath(name), v)
	return err
}

func (c *Client) ModLock_RetrieveValue(name string) (string, error) {
	raw, err := c.sendRequest("GET", nameToPath(name), nil)
	if err != nil {
		return "", err
	}
	return string(raw.Body), nil
}

func (c *Client) ModLock_RetrieveIndex(name string) (uint64, error) {
	raw, err := c.sendRequest("GET", nameToPath(name)+"?field=index", nil)
	if err != nil {
		return 0, err
	}
	index, err := strconv.ParseUint(string(raw.Body), 10, 0)
	return index, err
}

func (c *Client) ModLock_ReleaseWithIndex(name string, index uint64) error {
	_, err := c.sendRequest("DELETE", fmt.Sprintf("%s?index=%d", nameToPath(name), index), nil)
	return err
}

func (c *Client) ModLock_ReleaseWithValue(name string, value string) error {
	_, err := c.sendRequest("DELETE", fmt.Sprintf("%s?value=%s", nameToPath(name), value), nil)
	return err
}
