package etcd

// SetDir sets the given key to a directory.
func (c *Client) SetDir(key string, ttl uint64) (*Response, error) {

	raw, err := c.put(key, "", ttl, nil)

	if err != nil {
		return nil, err
	}

	return raw.toResponse()
}

// UpdateDir updates the given key to a directory.  It succeeds only if the
// given key already exists.
func (c *Client) UpdateDir(key string, ttl uint64) (*Response, error) {
	ops := options{
		"prevExist": true,
	}

	raw, err := c.put(key, "", ttl, ops)

	if err != nil {
		return nil, err
	}

	return raw.toResponse()
}

// UpdateDir creates a directory under the given key.  It succeeds only if
// the given key does not yet exist.
func (c *Client) CreateDir(key string, ttl uint64) (*Response, error) {
	ops := options{
		"prevExist": false,
	}

	raw, err := c.put(key, "", ttl, ops)

	if err != nil {
		return nil, err
	}

	return raw.toResponse()
}

// Set sets the given key to the given value.
func (c *Client) Set(key string, value string, ttl uint64) (*Response, error) {
	raw, err := c.put(key, value, ttl, nil)

	if err != nil {
		return nil, err
	}

	return raw.toResponse()
}

// Update updates the given key to the given value.  It succeeds only if the
// given key already exists.
func (c *Client) Update(key string, value string, ttl uint64) (*Response, error) {
	ops := options{
		"prevExist": true,
	}

	raw, err := c.put(key, value, ttl, ops)

	if err != nil {
		return nil, err
	}

	return raw.toResponse()
}

// Create creates a file with the given value under the given key.  It succeeds
// only if the given key does not yet exist.
func (c *Client) Create(key string, value string, ttl uint64) (*Response, error) {
	ops := options{
		"prevExist": false,
	}

	raw, err := c.put(key, value, ttl, ops)

	if err != nil {
		return nil, err
	}

	return raw.toResponse()
}
