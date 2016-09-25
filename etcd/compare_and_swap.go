package etcd

import "fmt"

func (c *Client) CompareAndSwap(key string, value string, ttl uint64,
	prevValue string, prevIndex uint64) (*Response, error) {
	raw, err := c.RawCompareAndSwap(key, value, ttl, prevValue, prevIndex, nil)
	if err != nil {
		return nil, err
	}

	return raw.Unmarshal()
}

func (c *Client) RawCompareAndSwap(key string, value string, ttl uint64,
	prevValue string, prevIndex uint64, prevExist *bool) (*RawResponse, error) {
	if prevValue == "" && prevIndex == 0 && prevExist == nil {
		return nil, fmt.Errorf("You must give either prevValue, prevIndex or prevExist.")
	}

	options := Options{}
	if prevValue != "" {
		options["prevValue"] = prevValue
	}
	if prevIndex != 0 {
		options["prevIndex"] = prevIndex
	}
	if prevExist != nil {
		if *prevExist {
			options["prevExist"] = "true"
		} else {
			options["prevExist"] = "false"
		}
	}

	raw, err := c.put(key, value, ttl, options)

	if err != nil {
		return nil, err
	}

	return raw, err
}
