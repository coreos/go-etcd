package etcd

import (
	"net/url"
)

// Lock uses the etcd lock module and locks the given path for the ttl
func (c *Client) Lock(key, value string, ttl uint64) (lockKey string, err error) {
	logger.Debugf("lock %s, %s, ttl: %d, [%s]", key, value, ttl, c.cluster.Leader)

	values := buildValues(value, ttl)
	req := NewRawModuleRequest("POST", key, values, ModuleLock, nil)
	rawResp, err := c.SendRequest(req)

	if err != nil {
		return "", err
	}

	return string(rawResp.Body), nil
}

// Lock uses the etcd lock module and locks the given path for the ttl
func (c *Client) LockRenew(key, value string, ttl uint64, lockKey string) (err error) {
	logger.Debugf("lock %s: %s, %s, ttl: %d, [%s]", key, lockKey, value, ttl, c.cluster.Leader)

	values := buildValues(value, ttl)
	values.Set("index", lockKey)
	req := NewRawModuleRequest("PUT", key, values, ModuleLock, nil)
	_, err = c.SendRequest(req)
	return err
}

// LockDelete removes an existing Lock.
func (c *Client) LockDelete(key, value string, lockKey string) (err error) {
	logger.Debugf("unlock %s: %s, %s, [%s]", key, lockKey, value, c.cluster.Leader)

	values := buildValues(value, 0)
	if len(lockKey) > 0 {
		values.Set("index", lockKey)
	}
	req := NewRawModuleRequest("DELETE", key, values, ModuleLock, nil)
	_, err = c.SendRequest(req)
	return err
}

// LockValue gets the current value of a lock, if it exists.
func (c *Client) LockValue(key, lockKey string) (value string, err error) {
	logger.Debugf("lock value %s: %s, [%s]", key, lockKey, c.cluster.Leader)

	values := url.Values{}
	values.Set("index", lockKey)
	req := NewRawModuleRequest("GET", key, values, ModuleLock, nil)
	rawResp, err := c.SendRequest(req)
	if err != nil {
		return "", err
	}
	return string(rawResp.Body), nil
}
