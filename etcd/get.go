package etcd

// Get gets the file or directory associated with the given key.
// If the key points to a directory, files and directories under
// it will be returned in sorted or unsorted order, depending on
// the sort flag.  Note that contents under child directories
// will not be returned.

// Set recursive to true, to get those contents, use GetAll.
func (c *Client) Get(key string, sort, recursive bool) (*Response, error) {
	ops := options{
		"recursive": true,
		"sorted":    sort,
	}

	r, err := c.get(key, ops, normalResponse)

	if err != nil {
		return nil, err
	}

	resp := r.(*Response)

	return resp, nil
}
