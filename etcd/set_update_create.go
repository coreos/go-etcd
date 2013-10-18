package etcd

func (c *Client) SetDir(key string, ttl uint64) (*Response, error) {
	return c.put(key, "", ttl, nil)
}

func (c *Client) UpdateDir(key string, ttl uint64) (*Response, error) {
	return c.put(key, "", ttl, Options{
		"prevExist": true,
	})
}

func (c *Client) CreateDir(key string, ttl uint64) (*Response, error) {
	return c.put(key, "", ttl, Options{
		"prevExist": false,
	})
}

func (c *Client) Set(key string, value string, ttl uint64) (*Response, error) {
	return c.put(key, value, ttl, nil)
}

func (c *Client) Update(key string, value string, ttl uint64) (*Response, error) {
	return c.put(key, value, ttl, Options{
		"prevExist": true,
	})
}

func (c *Client) Create(key string, value string, ttl uint64) (*Response, error) {
	return c.put(key, value, ttl, Options{
		"prevExist": false,
	})
}
