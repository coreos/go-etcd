package etcd

// Put issues a PUT request
func (c *Client) Put(key string, value string, ttl uint64,
	options Options) (*Response, error) {
	raw, err := c.put(key, value, ttl, options)
	if err != nil {
		return nil, err
	}

	return raw.Unmarshal()
}
