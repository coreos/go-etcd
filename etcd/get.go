package etcd

// GetDir gets the content of the given directory.
// If recursive is true, then the content of all child directories
// will also be returned.
// If sort is true, then the files will be sorted.
func (c *Client) GetDir(key string, recursive, sort bool) (*Response, error) {
	return c.get(key, options{
		"recursive": recursive,
		"sorted":    sort,
	})
}

// Get gets the value of the given key.
func (c *Client) Get(key string) (*Response, error) {
	return c.get(key, nil)
}
