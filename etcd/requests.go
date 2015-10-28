package etcd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

// Errors introduced by handling requests
var (
	ErrRequestCancelled = errors.New("sending request is cancelled")
)

type RawRequest struct {
	Method       string
	RelativePath string
	Values       url.Values
	Cancel       <-chan bool
}

// NewRawRequest returns a new RawRequest
func NewRawRequest(method, relativePath string, values url.Values, cancel <-chan bool) *RawRequest {
	return &RawRequest{
		Method:       method,
		RelativePath: relativePath,
		Values:       values,
		Cancel:       cancel,
	}
}

// getRequest constructs a RawRequest from the given parameters.
func (c *Client) getRequest(key string, options Options, cancel <-chan bool) (*RawRequest, error) {
	p := keyToPath(key)

	str, err := options.toParameters(VALID_GET_OPTIONS)
	if err != nil {
		return nil, err
	}
	p += str

	return NewRawRequest("GET", p, nil, cancel), nil
}

// getCancelable issues a cancelable GET request
func (c *Client) getCancelable(key string, options Options,
	cancel <-chan bool) (*RawResponse, error) {
	logger.Debugf("get %s [%s]", key, c.cluster.pick())

	req, err := c.getRequest(key, options, cancel)
	if err != nil {
		return nil, err
	}
	resp, err := c.SendRequest(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// get issues a GET request
func (c *Client) get(key string, options Options) (*RawResponse, error) {
	return c.getCancelable(key, options, nil)
}

// put issues a PUT request
func (c *Client) put(key string, value string, ttl uint64,
	options Options) (*RawResponse, error) {

	logger.Debugf("put %s, %s, ttl: %d, [%s]", key, value, ttl, c.cluster.pick())
	p := keyToPath(key)

	str, err := options.toParameters(VALID_PUT_OPTIONS)
	if err != nil {
		return nil, err
	}
	p += str

	req := NewRawRequest("PUT", p, buildValues(value, ttl), nil)
	resp, err := c.SendRequest(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// post issues a POST request
func (c *Client) post(key string, value string, ttl uint64) (*RawResponse, error) {
	logger.Debugf("post %s, %s, ttl: %d, [%s]", key, value, ttl, c.cluster.pick())
	p := keyToPath(key)

	req := NewRawRequest("POST", p, buildValues(value, ttl), nil)
	resp, err := c.SendRequest(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// delete issues a DELETE request
func (c *Client) delete(key string, options Options) (*RawResponse, error) {
	logger.Debugf("delete %s [%s]", key, c.cluster.pick())
	p := keyToPath(key)

	str, err := options.toParameters(VALID_DELETE_OPTIONS)
	if err != nil {
		return nil, err
	}
	p += str

	req := NewRawRequest("DELETE", p, nil, nil)
	resp, err := c.SendRequest(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// SendStreamRequest performs a watch that streams back results over the receiver channel.
func (c *Client) SendStreamRequest(rr *RawRequest, receiver chan *Response, errChan chan error) error {
	requestManager := NewRequestManager(rr.Cancel, c)
	if rr.Cancel != nil {
		defer close(requestManager.stopCh)
		go requestManager.CancellationHandler()
	}
	resp, err := requestManager.send(rr)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return err
	}
	decoder := json.NewDecoder(resp.Body)
	go func() {
		defer resp.Body.Close()
		for {
			got := new(Response)
			err = decoder.Decode(got)

			// Decode error may be caused due to cancel request
			select {
			case <-requestManager.CancellationNotifyChan:
				errChan <- ErrRequestCancelled
				return
			default:
			}
			if err != nil {
				errChan <- err
				return
			}
			// TODO: Dig up if these headers are consistent across stream
			addHeaders(got, resp.Header)
			receiver <- got
		}
	}()
	return nil
}

// SendRequest sends a HTTP request and returns a Response as defined by etcd
func (c *Client) SendRequest(rr *RawRequest) (*RawResponse, error) {

	requestManager := NewRequestManager(rr.Cancel, c)
	if rr.Cancel != nil {
		defer close(requestManager.stopCh)
		go requestManager.CancellationHandler()
	}

	resp, err := requestManager.send(rr)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	// try to read byte code and break the loop
	respBody, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		// TODO: Log httpPath
		logger.Debug("recv.success ")
	} else if err == io.ErrUnexpectedEOF {
		// underlying connection was closed prematurely, probably by timeout
		// TODO: empty body or unexpectedEOF can cause http.Transport to get hosed;
		// this allows the client to detect that and take evasive action. Need
		// to revisit once code.google.com/p/go/issues/detail?id=8648 gets fixed.
		respBody = []byte{}
	} else {
		// ReadAll error may be caused due to cancel request
		select {
		case <-requestManager.CancellationNotifyChan:
			return nil, ErrRequestCancelled
		default:
		}
	}
	r := &RawResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Header:     resp.Header,
	}
	return r, nil
}

type requestManager struct {
	// Request sync lock
	reqLock sync.Mutex
	// Request to cancel
	req *http.Request
	// Channel to receive cancellation notice
	CancellationReceiveChan <-chan bool
	// Channel to send cancellation notification
	CancellationNotifyChan chan bool
	// Channel to receive stop signals
	stopCh chan bool
	// http client
	c *Client
}

func NewRequestManager(cancellationReceiveChan <-chan bool, c *Client) *requestManager {
	return &requestManager{sync.Mutex{}, &http.Request{}, cancellationReceiveChan, make(chan bool, 1), make(chan bool), c}
}

func (r *requestManager) CancellationHandler() {
	select {
	case <-r.CancellationReceiveChan:
		r.CancellationNotifyChan <- true
		logger.Debug("send.request is cancelled")
	case <-r.stopCh:
		return
	}

	// Repeat canceling request until this thread is stopped
	// because we have no idea about whether it succeeds.
	for {
		r.reqLock.Lock()
		r.c.httpClient.Transport.(*http.Transport).CancelRequest(r.req)
		r.reqLock.Unlock()

		select {
		case <-time.After(100 * time.Millisecond):
		case <-r.stopCh:
			return
		}
	}
}

// send sends an HTTP request and returns an http.Response. It's up to the caller to close the response, or setup cancellation.
func (r *requestManager) send(rr *RawRequest) (*http.Response, error) {
	var resp *http.Response
	var httpPath string
	var err error

	var numReqs = 1

	logger.Debug("requests: SendRequest %+v", rr)
	checkRetry := r.c.CheckRetry
	if checkRetry == nil {
		checkRetry = DefaultCheckRetry
	}

	// If we connect to a follower and consistency is required, retry until
	// we connect to a leader
	sleep := 25 * time.Millisecond
	maxSleep := time.Second

	for attempt := 0; ; attempt++ {
		if attempt > 0 {
			select {
			case <-r.CancellationNotifyChan:
				return nil, ErrRequestCancelled
			case <-time.After(sleep):
				sleep = sleep * 2
				if sleep > maxSleep {
					sleep = maxSleep
				}
			}
		}

		logger.Debug("Connecting to etcd: attempt ", attempt+1, " for ", rr.RelativePath)

		// get httpPath if not set
		if httpPath == "" {
			httpPath = r.c.getHttpPath(rr.RelativePath)
		}

		// Return a cURL command if curlChan is set
		if r.c.cURLch != nil {
			command := fmt.Sprintf("curl -X %s %s", rr.Method, httpPath)
			for key, value := range rr.Values {
				command += fmt.Sprintf(" -d %s=%s", key, value[0])
			}
			if r.c.credentials != nil {
				command += fmt.Sprintf(" -u %s", r.c.credentials.username)
			}
			r.c.sendCURL(command)
		}

		logger.Debug("send.request.to ", httpPath, " | method ", rr.Method)

		err := func() error {
			// Note that this request is synchronized with the requestManager's cancellation watch dog
			r.reqLock.Lock()
			defer r.reqLock.Unlock()

			if rr.Values == nil {
				if r.req, err = http.NewRequest(rr.Method, httpPath, nil); err != nil {
					return err
				}
			} else {
				body := strings.NewReader(rr.Values.Encode())
				if r.req, err = http.NewRequest(rr.Method, httpPath, body); err != nil {
					return err
				}

				r.req.Header.Set("Content-Type",
					"application/x-www-form-urlencoded; param=value")
			}
			return nil
		}()

		if err != nil {
			return nil, err
		}

		if r.c.credentials != nil {
			r.req.SetBasicAuth(r.c.credentials.username, r.c.credentials.password)
		}

		resp, err = r.c.httpClient.Do(r.req)
		// clear previous httpPath
		httpPath = ""

		// If the request was cancelled, return ErrRequestCancelled directly
		select {
		case <-r.CancellationNotifyChan:
			return nil, ErrRequestCancelled
		default:
		}

		numReqs++

		// network error, change a machine!
		if err != nil {
			logger.Debug("network error: ", err.Error())
			lastResp := http.Response{}
			if checkErr := checkRetry(r.c.cluster, numReqs, lastResp, err); checkErr != nil {
				return nil, checkErr
			}

			r.c.cluster.failure()
			continue
		}

		// if there is no error, it should receive response
		logger.Debug("recv.response.from ", httpPath)

		if validHttpStatusCode[resp.StatusCode] {
			return resp, nil
		}

		if resp.StatusCode == http.StatusTemporaryRedirect {
			u, err := resp.Location()

			if err != nil {
				logger.Warning(err)
			} else {
				// set httpPath for following redirection
				httpPath = u.String()
			}
			// redirects don't count toward retries
			resp.Body.Close()
			continue
		}

		// We're either going to retry, or fail having exceeded retries. Eitherway we need to close this response.
		resp.Body.Close()

		if checkErr := checkRetry(r.c.cluster, numReqs, *resp,
			errors.New("Unexpected HTTP status code")); checkErr != nil {
			return nil, checkErr
		}
	}
}

// DefaultCheckRetry defines the retrying behaviour for bad HTTP requests
// If we have retried 2 * machine number, stop retrying.
// If status code is InternalServerError, sleep for 200ms.
func DefaultCheckRetry(cluster *Cluster, numReqs int, lastResp http.Response,
	err error) error {

	if numReqs > 2*len(cluster.Machines) {
		errStr := fmt.Sprintf("failed to propose on members %v twice [last error: %v]", cluster.Machines, err)
		return newError(ErrCodeEtcdNotReachable, errStr, 0)
	}

	if isEmptyResponse(lastResp) {
		// always retry if it failed to get response from one machine
		return nil
	}
	if !shouldRetry(lastResp) {
		body := []byte("nil")
		if lastResp.Body != nil {
			if b, err := ioutil.ReadAll(lastResp.Body); err == nil {
				body = b
			}
		}
		errStr := fmt.Sprintf("unhandled http status [%s] with body [%s]", http.StatusText(lastResp.StatusCode), body)
		return newError(ErrCodeUnhandledHTTPStatus, errStr, 0)
	}
	// sleep some time and expect leader election finish
	time.Sleep(time.Millisecond * 200)
	logger.Warning("bad response status code ", lastResp.StatusCode)
	return nil
}

func isEmptyResponse(r http.Response) bool { return r.StatusCode == 0 }

// shouldRetry returns whether the reponse deserves retry.
func shouldRetry(r http.Response) bool {
	// TODO: only retry when the cluster is in leader election
	// We cannot do it exactly because etcd doesn't support it well.
	return r.StatusCode == http.StatusInternalServerError
}

func (c *Client) getHttpPath(s ...string) string {
	fullPath := c.cluster.pick() + "/" + version
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

// convert key string to http path exclude version, including URL escaping
// for example: key[foo] -> path[keys/foo]
// key[/%z] -> path[keys/%25z]
// key[/] -> path[keys/]
func keyToPath(key string) string {
	// URL-escape our key, except for slashes
	p := strings.Replace(url.QueryEscape(path.Join("keys", key)), "%2F", "/", -1)

	// corner case: if key is "/" or "//" ect
	// path join will clear the tailing "/"
	// we need to add it back
	if p == "keys" {
		p = "keys/"
	}

	return p
}
