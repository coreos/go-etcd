package etcd

func (c *Client) GetDir(key string, sort bool) (*Response, error) {
	return c.get(key, options{
		"recursive": true,
		"sorted":    sort,
	})
}

func (c *Client) Get(key string, sort bool) (*Response, error) {
	return c.get(key, options{
		"sorted": sort,
	})
}
