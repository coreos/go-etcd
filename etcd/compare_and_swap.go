package etcd

import "fmt"

func (c *Client) CompareAndSwap(key string, value string, ttl uint64, prevValue string, prevIndex uint64) (*Response, error) {
	if prevValue == "" && prevIndex == 0 {
		return nil, fmt.Errorf("You must give either prevValue or prevIndex.")
	}

	options := options{}
	if prevValue != "" {
		options["prevValue"] = prevValue
	}
	if prevIndex != 0 {
		options["prevIndex"] = prevIndex
	}

	return toResp(c.put(key, value, ttl, options, normalResponse))
}
