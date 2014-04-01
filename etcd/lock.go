package etcd

import (
	"errors"
)

var (
	ErrInvalidLockKey = errors.New("go-etcd: invalid lock key")
)

// Lock uses the etcd lock module and locks the given path for the ttl
func (c *Client) Lock(key, value string, ttl uint64) (lockKey string, err error) {

	if len(key) == 0 || len(key) < 1 {
		return "", ErrInvalidLockKey
	}

	logger.Debugf("lock %s, %s, ttl: %d, [%s]", key, value, ttl, c.cluster.Leader)

	values := buildValues(value, ttl)
	req := NewRawRequest("POST", key, values, nil)
	rawResp, err := c.SendRequest(req, MODULE_LOCK)

	if err != nil {
		return "", err
	}

	return string(rawResp.Body), nil
}

// Lock uses the etcd lock module and locks the given path for the ttl
func (c *Client) LockRenew(key, value string, ttl uint64, lockKey string) (err error) {
	if len(key) == 0 || len(key) < 1 {
		return ErrInvalidLockKey
	}

	logger.Debugf("lock %s, %s, ttl: %d, [%s]", key, value, ttl, c.cluster.Leader)

	values := buildValues(value, ttl)
	values.Set("index", lockKey)
	req := NewRawRequest("PUT", key, values, nil)
	_, err = c.SendRequest(req, MODULE_LOCK)
	return err
}

// LockRemove removes an existing Lock.
func (c *Client) LockRemove(key, value string, lockKey string) (err error) {
	if len(key) == 0 || len(key) < 1 {
		return ErrInvalidLockKey
	}

	logger.Debugf("unlock %s, %s, [%s]", key, value, c.cluster.Leader)

	values := buildValues(value, 0)
	values.Set("index", lockKey)
	req := NewRawRequest("DELETE", key, values, nil)
	_, err = c.SendRequest(req, MODULE_LOCK)
	return err
}
