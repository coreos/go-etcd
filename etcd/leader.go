package etcd

import (
	"net/url"
)

// LeaderLock uses the etcd lock module and locks the given path for the ttl. LeaderLock
// may be called multiple time to renew the lock.
func (c *Client) LeaderLock(key, name string, ttl uint64) (leaderKey string, err error) {

	values := buildValues("", ttl)
	values.Set("name", name)
	req := NewRawModuleRequest("PUT", key, values, ModuleLeader, nil)
	rawResp, err := c.SendRequest(req)

	if err != nil {
		return "", err
	}

	return string(rawResp.Body), nil
}

// LeaderValue retrieves the current name for the leader by key
func (c *Client) LeaderValue(key string) (value string, err error) {
	req := NewRawModuleRequest("GET", key, url.Values{}, ModuleLeader, nil)
	rawResp, err := c.SendRequest(req)

	if err != nil {
		return "", err
	}

	return string(rawResp.Body), nil
}

// LeaderDelete deletes a leader for key by name
func (c *Client) LeaderDelete(key, name string) (err error) {
	values := url.Values{}
	values.Set("name", name)
	req := NewRawModuleRequest("GET", key, values, ModuleLeader, nil)
	_, err = c.SendRequest(req)
	return err
}
