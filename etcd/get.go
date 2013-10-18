package etcd

func (c *Client) GetDir(key string, sort bool) (*Response, error) {
	return c.get(key, Options{
		"recursive": true,
		"sorted":    sort,
	})
}

func (c *Client) Get(key string, sort bool) (*Response, error) {
	return c.get(key, Options{
		"sorted": sort,
	})
}
