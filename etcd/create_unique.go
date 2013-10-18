package etcd

func (c *Client) CreateUniqueDir(key string, ttl uint64) (*Response, error) {
	return c.post(key, "", ttl)
}

func (c *Client) CreateUnique(key string, value string, ttl uint64) (*Response, error) {
	return c.post(key, value, ttl)
}
