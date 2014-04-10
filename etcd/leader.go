package etcd

import (
	"net/url"
)

// Leader is type to manage leaders created with the etcd leader module.
type Leader struct {
	c    *Client // go-etcd client
	key  string  // key of leader
	name string  // name of leader
}

// NewLeader creates a Leader object created with the etcd leader module.
func (c *Client) NewLeader(key, name string) (l *Leader) {
	return &Leader{
		c:    c,
		key:  key,
		name: name,
	}
}

// Key returns the key for the leader object.
func (l *Leader) Key() string {
	return l.key
}

// Name returns the name for the leader object.
func (l *Leader) Name() string {
	return l.name
}

// Lead attempts to aquire the leader lock. It will block if another client has
// already locked it. Lead can be called again to renew the Leader role.
func (l *Leader) Lead(ttl uint64) (err error) {

	values := buildValues("", ttl)
	values.Set("name", l.name)
	req := NewRawModuleRequest("PUT", l.key, values, ModuleLeader, nil)
	_, err = l.c.SendRequest(req)

	return err
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
