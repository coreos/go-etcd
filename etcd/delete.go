package etcd

func (c *Client) DeleteDir(key string) (*Response, error) {
	return c.delete(key, Options{
		"recursive": true,
	})
}

func (c *Client) Delete(key string) (*Response, error) {
	return c.delete(key, nil)
}
