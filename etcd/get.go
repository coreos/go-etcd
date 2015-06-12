package etcd

// Get gets the file or directory associated with the given key.
// If the key points to a directory, files and directories under
// it will be returned in sorted or unsorted order, depending on
// the sort flag.
// If recursive is set to false, contents under child directories
// will not be returned.
// If recursive is set to true, all the contents will be returned.
func (c *Client) Get(key string, sort, recursive bool) (*Response, error) {
	raw, err := c.RawGet(key, sort, recursive, false)

	if err != nil {
		return nil, err
	}

	return raw.Unmarshal()
}
// aka GET with 'quorum=true'
func (c *Client) LinearizedGet(key string, sort, recursive bool) (*Response, error) {
	raw, err := c.RawGet(key, sort, recursive, true)

	if err != nil {
		return nil, err
	}

	return raw.Unmarshal()
}

func (c *Client) RawGet(key string, sort, recursive bool, quorum bool) (*RawResponse, error) {
	q := (c.config.Consistency == STRONG_CONSISTENCY) || quorum

	ops := Options{
		"recursive": recursive,
		"sorted":    sort,
		"quorum":    q,
	}

	return c.get(key, ops)
}
