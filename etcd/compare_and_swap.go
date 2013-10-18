package etcd

func (c *Client) CompareAndSwap(key string, value string, ttl uint64, options Options) (*Response, error) {
	return c.put(key, value, ttl, options)
}
