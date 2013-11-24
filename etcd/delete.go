package etcd

// Delete deletes the given key.
// When recursive set to false If the key points to a
// directory, the method will fail.
// When recursive set to true, if the key points to a file,
// the file will be deleted.  If the key points
// to a directory, then everything under the directory, include
// all child directories, will be deleted.
func (c *Client) Delete(key string, recursive bool) (*Response, error) {
	ops := options{
		"recursive": recursive,
	}

	return toResp(c.delete(key, ops, normalResponse))
}
