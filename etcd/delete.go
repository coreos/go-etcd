package etcd

// Delete the given directory and everything under it.
func (c *Client) DeleteDir(key string) (*Response, error) {
	return c.delete(key, options{
		"recursive": true,
	})
}

// Delete the given key.
func (c *Client) Delete(key string) (*Response, error) {
	return c.delete(key, nil)
}
