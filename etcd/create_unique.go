package etcd

// Create a new directory with a random etcd-generated key under the given path.
func (c *Client) CreateUniqueDir(key string, ttl uint64) (*Response, error) {
	return c.post(key, "", ttl)
}

// Create a new file with a random etcd-generated key under the given path.
func (c *Client) CreateUnique(key string, value string, ttl uint64) (*Response, error) {
	return c.post(key, value, ttl)
}
