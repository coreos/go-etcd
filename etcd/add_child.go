package etcd

// Add a new directory with a random etcd-generated key under the given path.
func (c *Client) AddChildDir(key string, ttl uint64) (*Response, error) {
	return toResp(c.post(key, "", ttl, normalResponse))
}

// Add a new file with a random etcd-generated key under the given path.
func (c *Client) AddChild(key string, value string, ttl uint64) (*Response, error) {
	return toResp(c.post(key, value, ttl, normalResponse))
}
