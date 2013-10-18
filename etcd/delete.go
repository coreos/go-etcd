package etcd

func (c *Client) DeleteDir(key string) (*Response, error) {
	return c.delete(key, options{
		"recursive": true,
	})
}

func (c *Client) Delete(key string) (*Response, error) {
	return c.delete(key, nil)
}
