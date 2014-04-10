package etcd

import (
	"net/url"
)

// Leader is type to manage leaders created with the etcd leader module.
type Leader struct {
	c       *Client // go-etcd client
	key     string  // key of leader
	name    string  // name of leader
	ttl     uint64  // time to live of leader
	lockKey string  // index given to leader instance by etcd
}

// NewLeader creates a Leader object created with the etcd leader module.
func (c *Client) NewLeader(key, name string, ttl uint64) (l *Leader) {
	return &Leader{
		c:    c,
		key:  key,
		name: name,
		ttl:  ttl,
	}
}

// TTL gets the ttl in seconds
func (l *Leader) TTL() uint64 {
	return l.ttl
}

// SetTTL sets the ttl in seconds
func (l *Leader) SetTTL(ttl uint64) {
	l.ttl = ttl
}

// Lock attempts to aquire the leader lock. It will block if another client has
// already locked it.
func (l *Leader) Lock() (err error) {

	values := buildValues("", l.ttl)
	values.Set("name", l.name)
	req := NewRawModuleRequest("PUT", l.key, values, ModuleLeader, nil)
	rawResp, err := l.c.SendRequest(req)

	if err != nil {
		return err
	}

	l.lockKey = string(rawResp.Body)
	return nil
}

// Current retrieves the name for the current leader
func (l *Leader) Current() (value string, err error) {
	req := NewRawModuleRequest("GET", l.key, url.Values{}, ModuleLeader, nil)
	rawResp, err := l.c.SendRequest(req)

	if err != nil {
		return "", err
	}

	return string(rawResp.Body), nil
}

// LeaderDelete deletes a leader
func (l *Leader) Delete() (err error) {
	values := url.Values{}
	values.Set("name", l.name)
	req := NewRawModuleRequest("GET", l.key, values, ModuleLeader, nil)
	_, err = l.c.SendRequest(req)
	return err
}
