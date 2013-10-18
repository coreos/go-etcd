package etcd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"
	"reflect"
)

var (
	VALID_GET_OPTIONS = validOptions{
		"recursive":  reflect.Bool,
		"consistent": reflect.Bool,
		"sorted":     reflect.Bool,
		"wait":       reflect.Bool,
		"wait_index": reflect.Int,
	}
)

func (c *Client) GetDir(key string, options Options) ([]*Response, error) {
	options["recursive"] = true
	return c.Get(key, options)
}

func (c *Client) Get(key string, options Options) ([]*Response, error) {
	logger.Debugf("get %s [%s]", key, c.cluster.Leader)

	p := path.Join("keys", key)
	if options != nil {
		str, err := optionsToString(options, VALID_GET_OPTIONS)
		if err != nil {
			return nil, err
		}
		p += str
	}

	resp, err := c.sendRequest("GET", p, "")

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

	return convertGetResponse(b)
}

func (c *Client) GetDirFrom(key string, addr string, options Options) ([]*Response, error) {
	options["recursive"] = true
	return c.GetFrom(key, addr, options)
}

// GetFrom gets the value of the key from a given machine address.
// If the given machine is not available it returns an error.
// Mainly use for testing purpose
func (c *Client) GetFrom(key string, addr string, options Options) ([]*Response, error) {

	httpPath := c.createHttpPath(addr, path.Join(version, "keys", key))

	if options != nil {
		str, err := optionsToString(options, VALID_GET_OPTIONS)
		if err != nil {
			return nil, err
		}
		httpPath += str
	}

	resp, err := c.httpClient.Get(httpPath)

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

	return convertGetResponse(b)
}

// Convert byte stream to response.
func convertGetResponse(b []byte) ([]*Response, error) {

	var results []*Response
	var result *Response

	err := json.Unmarshal(b, &result)

	if err != nil {
		err = json.Unmarshal(b, &results)

		if err != nil {
			return nil, err
		}

	} else {
		results = make([]*Response, 1)
		results[0] = result
	}
	return results, nil
}
