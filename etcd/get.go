package etcd

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/store"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
)

type Options map[string]interface{}

var (
	// Making a map to make it easier to test existence
	validGetOptions = map[string]bool{
		"recursive":  true,
		"consistent": true,
		"sorted":     true,
		"wait":       true,
		"wait_index": true,
	}
)

// Get the value of the given key
func (c *Client) Get(key string) ([]*store.Response, error) {
	return c.GetWithOptions(key, nil)
}

// The same with Get, but allows passing options
func (c *Client) GetWithOptions(key string, options Options) ([]*store.Response, error) {
	logger.Debugf("get %s [%s]", key, c.cluster.Leader)

	p := path.Join("keys", key)
	if options != nil {
		str, err := optionsToString(options)
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

// GetTo gets the value of the key from a given machine address.
// If the given machine is not available it returns an error.
// Mainly use for testing purpose
func (c *Client) GetFrom(key string, addr string) ([]*store.Response, error) {
	return c.GetFromWithOptions(key, addr, nil)
}

// The same with GetFrom, but allows passing options
func (c *Client) GetFromWithOptions(key string, addr string, options Options) ([]*store.Response, error) {
	httpPath := c.createHttpPath(addr, path.Join(version, "keys", key))

	if options != nil {
		str, err := optionsToString(options)
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
func convertGetResponse(b []byte) ([]*store.Response, error) {

	var results []*store.Response
	var result *store.Response

	err := json.Unmarshal(b, &result)

	if err != nil {
		err = json.Unmarshal(b, &results)

		if err != nil {
			return nil, err
		}

	} else {
		results = make([]*store.Response, 1)
		results[0] = result
	}
	return results, nil
}

// Convert options to a string of HTML parameters
func optionsToString(options Options) (string, error) {
	p := "?"
	optionArr := make([]string, 0)
	for opKey, opVal := range options {
		if validGetOptions[opKey] {
			optionArr = append(optionArr, fmt.Sprintf("%v=%v", opKey, opVal))
		} else {
			return "", fmt.Errorf("Invalid option: %v", opKey)
		}
	}
	p += strings.Join(optionArr, "&")
	return p, nil
}
