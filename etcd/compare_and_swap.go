package etcd

import "fmt"

type PrevExist int
const ( 
    Ignored = iota
    Exists
    DoesNotExist
)

func (c *Client) CompareAndSwap(key string, value string, ttl uint64,
	prevValue string, prevIndex uint64, prevExist PrevExist) (*Response, error) {
	raw, err := c.RawCompareAndSwap(key, value, ttl, prevValue, prevIndex, prevExist)
	if err != nil {
		return nil, err
	}

	return raw.Unmarshal()
}

func (c *Client) RawCompareAndSwap(key string, value string, ttl uint64,
	prevValue string, prevIndex uint64, prevExist PrevExist) (*RawResponse, error) {
	if prevValue == "" && prevIndex == 0 && prevExist == Ignored {
		return nil, fmt.Errorf("You must give either prevValue, prevIndex or prevExist.")
	}

	options := Options{}
	if prevValue != "" {
		options["prevValue"] = prevValue
	}
	if prevIndex != 0 {
		options["prevIndex"] = prevIndex
	}
	if prevExist != Ignored {
		options["prevExist"] = prevExist == Exists
	}

	raw, err := c.put(key, value, ttl, options)

	if err != nil {
		return nil, err
	}

	return raw, err
}
