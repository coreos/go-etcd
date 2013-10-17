package etcd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"reflect"
)

var (
	VALID_TESTANDSET_OPTIONS = validOptions{
		"prevValue": reflect.String,
		"prevIndex": reflect.Int,
		"prevExist": reflect.Bool,
	}
)

func (c *Client) TestAndSet(key string, value string, ttl uint64, options Options) (*Response, error) {
	// logger.Debugf("set %s, %s[%s], ttl: %d, [%s]", key, value, prevValue, ttl, c.cluster.Leader)
	v := url.Values{}
	v.Set("value", value)
	if ttl > 0 {
		v.Set("ttl", fmt.Sprintf("%v", ttl))
	}

	p := path.Join("keys", key)
	if options != nil {
		str, err := optionsToString(options, VALID_TESTANDSET_OPTIONS)
		if err != nil {
			return nil, err
		}
		p += str
	}

	resp, err := c.sendRequest("PUT", p, v.Encode())

	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)

	resp.Body.Close()

	if err != nil {

		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, handleError(b)
	}

	var result Response

	err = json.Unmarshal(b, &result)

	if err != nil {
		return nil, err
	}

	return &result, nil
}
