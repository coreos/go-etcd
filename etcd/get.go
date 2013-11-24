package etcd

// Get gets the file or directory associated with the given key.
// If the key points to a directory, files and directories under
// it will be returned in sorted or unsorted order, depending on
// the sort flag.  Note that contents under child directories
// will not be returned.
// Set recursive to true, to get those contents.

func (c *Client) Get(key string, sort, recursive bool) (*Response, error) {
	raw, err := c.RawGet(key, sort, recursive)

	if err != nil {
		return nil, err
	}

	return raw.toResponse()
}

func (c *Client) RawGet(key string, sort, recursive bool) (*RawResponse, error) {
	ops := options{
		"recursive": recursive,
		"sorted":    sort,
	}

	return c.get(key, ops)
}
