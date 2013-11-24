package etcd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

var (
	curlChan chan string
)

// SetCurlChan sets a channel to which cURL commands which can be used to
// re-produce requests are sent.  This is useful for debugging.
func SetCurlChan(c chan string) {
	curlChan = c
}

// get issues a GET request
func (c *Client) get(key string, options options, respType responseType) (interface{}, error) {
	logger.Debugf("get %s [%s]", key, c.cluster.Leader)

	p := path.Join("keys", key)
	// If consistency level is set to STRONG, append
	// the `consistent` query string.
	if c.config.Consistency == STRONG_CONSISTENCY {
		options["consistent"] = true
	}

	str, err := options.toParameters(VALID_GET_OPTIONS)
	if err != nil {
		return nil, err
	}
	p += str

	resp, err := c.sendRequest("GET", p, nil, respType)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// put issues a PUT request
func (c *Client) put(key string, value string, ttl uint64, options options,
	respType responseType) (interface{}, error) {

	logger.Debugf("put %s, %s, ttl: %d, [%s]", key, value, ttl, c.cluster.Leader)
	p := path.Join("keys", key)

	str, err := options.toParameters(VALID_PUT_OPTIONS)
	if err != nil {
		return nil, err
	}
	p += str

	resp, err := c.sendRequest("PUT", p, buildValues(value, ttl), respType)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// post issues a POST request
func (c *Client) post(key string, value string, ttl uint64, respType responseType) (interface{}, error) {
	logger.Debugf("post %s, %s, ttl: %d, [%s]", key, value, ttl, c.cluster.Leader)
	p := path.Join("keys", key)

	resp, err := c.sendRequest("POST", p, buildValues(value, ttl), respType)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// delete issues a DELETE request
func (c *Client) delete(key string, options options, respType responseType) (interface{}, error) {
	logger.Debugf("delete %s [%s]", key, c.cluster.Leader)

	p := path.Join("keys", key)

	str, err := options.toParameters(VALID_DELETE_OPTIONS)
	if err != nil {
		return nil, err
	}
	p += str

	resp, err := c.sendRequest("DELETE", p, nil, respType)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// sendRequest sends a HTTP request and returns a Response as defined by etcd
func (c *Client) sendRequest(method string, relativePath string, values url.Values,
	respType responseType) (interface{}, error) {

	var req *http.Request
	var resp *http.Response
	var httpPath string
	var err error
	var b []byte

	trial := 0

	// if we connect to a follower, we will retry until we found a leader
	for {
		trial++
		if trial > 2*len(c.cluster.Machines) {
			return nil, fmt.Errorf("Cannot reach servers after %v time", trial)
		}

		if method == "GET" && c.config.Consistency == WEAK_CONSISTENCY {
			// If it's a GET and consistency level is set to WEAK,
			// then use a random machine.
			httpPath = c.getHttpPath(true, relativePath)
		} else {
			// Else use the leader.
			httpPath = c.getHttpPath(false, relativePath)
		}

		// Return a cURL command if curlChan is set
		if curlChan != nil {
			command := fmt.Sprintf("curl -X %s %s", method, httpPath)
			for key, value := range values {
				command += fmt.Sprintf(" -d %s=%s", key, value[0])
			}
			curlChan <- command
		}

		logger.Debug("send.request.to ", httpPath, " | method ", method)

		if values == nil {
			req, _ = http.NewRequest(method, httpPath, nil)
		} else {
			req, _ = http.NewRequest(method, httpPath,
				strings.NewReader(values.Encode()))

			req.Header.Set("Content-Type",
				"application/x-www-form-urlencoded; param=value")
		}

		// network error, change a machine!
		if resp, err = c.httpClient.Do(req); err != nil {
			c.switchLeader(trial % len(c.cluster.Machines))
			time.Sleep(time.Millisecond * 200)
			continue
		}

		if resp != nil {
			logger.Debug("recv.response.from ", httpPath)

			var ok bool
			ok, b = c.handleResp(resp)

			if !ok {
				continue
			}

			logger.Debug("recv.success.", httpPath)
			break
		}

		// should not reach here
		// err and resp should not be nil at the same time
		logger.Debug("error.from ", httpPath)
		return nil, err
	}

	if respType == rawResponse {
		return &RawResponse{Body: b, Header: resp.Header}, nil
	}

	var result *Response

	err = json.Unmarshal(b, result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// handle HTTP response
// If status code is OK, read the http body.
// If status code is TemporaryRedirect, update leader.
// If status coid is InternalServerError, sleep for 200ms.
func (c *Client) handleResp(resp *http.Response) (bool, []byte) {
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTemporaryRedirect {
		u, err := resp.Location()

		if err != nil {
			logger.Warning(err)
		} else {
			c.updateLeader(u)
		}

	} else if resp.StatusCode == http.StatusInternalServerError {
		time.Sleep(time.Millisecond * 200)

	} else if resp.StatusCode == http.StatusOK ||
		resp.StatusCode == http.StatusCreated {
		b, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return false, nil
		}

		return true, b
	}

	logger.Warning("bad status code ", resp.StatusCode)
	return false, nil
}

func (c *Client) getHttpPath(random bool, s ...string) string {
	var machine string
	if random {
		machine = c.cluster.Machines[rand.Intn(len(c.cluster.Machines))]
	} else {
		machine = c.cluster.Leader
	}

	fullPath := machine + "/" + version
	for _, seg := range s {
		fullPath = fullPath + "/" + seg
	}

	return fullPath
}

// buildValues builds a url.Values map according to the given value and ttl
func buildValues(value string, ttl uint64) url.Values {
	v := url.Values{}

	if value != "" {
		v.Set("value", value)
	}

	if ttl > 0 {
		v.Set("ttl", fmt.Sprintf("%v", ttl))
	}

	return v
}
