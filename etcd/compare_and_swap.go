package etcd

func (c *Client) CompareAndSwap(key string, value string, ttl uint64, prevValue string, prevIndex uint64) (*Response, error) {
	options := options{}
	if prevValue != "" {
		options["prevValue"] = prevValue
	}
	if prevIndex != 0 {
		options["prevIndex"] = prevIndex
	}
	return c.put(key, value, ttl, options)
}
