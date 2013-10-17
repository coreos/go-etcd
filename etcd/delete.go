package etcd

import (
	"encoding/json"
	"github.com/coreos/etcd/store"
	"io/ioutil"
	"net/http"
	"path"
	"reflect"
)

const (
	VALID_DELETE_OPTIONS = validOptions{
		"recursive": reflect.Bool,
	}
)

func (c *Client) Delete(key string, options Options) (*store.Response, error) {

	p := path.Join("keys", key)
	if options != nil {
		str, err := optionsToString(options, VALID_DELETE_OPTIONS)
		if err != nil {
			return nil, err
		}
		p += str
	}

	resp, err := c.sendRequest("DELETE", p, "")

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

	var result store.Response

	err = json.Unmarshal(b, &result)

	if err != nil {
		return nil, err
	}

	return &result, nil

}
