package etcd

import (
	"net/url"
)

// Lock is type to manage locks created with the etcd lock module.
type Lock struct {
	c       *Client // go-etcd client
	key     string  // name of lock
	value   string  // value of lock
	ttl     uint64  // time to live of lock
	lockKey string  // index given to lock instance by etcd
}

// Lock uses the etcd lock module and locks the given path for the ttl
func (c *Client) NewLock(key, value string, ttl uint64) (l *Lock) {
	return &Lock{
		c:     c,
		key:   key,
		value: value,
		ttl:   ttl,
	}
}

// TTL gets the ttl in seconds
func (l *Lock) TTL() uint64 {
	return l.ttl
}

// SetTTL sets the ttl in seconds
func (l *Lock) SetTTL(ttl uint64) {
	l.ttl = ttl
}

func (l *Lock) Lock() (err error) {
	logger.Debugf("lock %s, %s, ttl: %d, [%s]", l.key, l.value, l.ttl, l.c.cluster.Leader)

	values := buildValues(l.value, l.ttl)
	req := NewRawModuleRequest("POST", l.key, values, ModuleLock, nil)
	rawResp, err := l.c.SendRequest(req)

	if err != nil {
		return err
	}

	l.lockKey = string(rawResp.Body)
	return nil
}

// Lock uses the etcd lock module and locks the given path for the ttl
func (l *Lock) Renew() (err error) {
	logger.Debugf("lock %s: %s, %s, ttl: %d, [%s]", l.key, l.lockKey, l.value, l.ttl, l.c.cluster.Leader)

	values := buildValues(l.value, l.ttl)
	values.Set("index", l.lockKey)
	req := NewRawModuleRequest("PUT", l.key, values, ModuleLock, nil)
	res, err := l.c.SendRequest(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		err = handleError(res.Body)
	}
	return err
}

// Current retrieves the current value for the lock
func (l *Lock) Current() (value string, err error) {
	req := NewRawModuleRequest("GET", l.key, url.Values{}, ModuleLock, nil)
	rawResp, err := l.c.SendRequest(req)

	if err != nil {
		return "", err
	}

	return string(rawResp.Body), nil
}

// Delete removes an existing Lock.
func (l *Lock) Delete() (err error) {
	logger.Debugf("unlock %s: %s, %s, [%s]", l.key, l.lockKey, l.value, l.c.cluster.Leader)

	values := url.Values{}
	values.Set("index", l.lockKey)
	req := NewRawModuleRequest("DELETE", l.key, values, ModuleLock, nil)
	res, err := l.c.SendRequest(req)
	if err != nil {
		return err
	}
	if res.StatusCode == 200 {
		return nil
	}
	return handleError(res.Body)
}
