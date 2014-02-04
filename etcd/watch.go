package etcd

import (
	"errors"
)

// Errors introduced by the Watch command.
var (
	ErrWatchStoppedByUser = errors.New("Watch stopped by the user via stop channel")
)

// If recursive is set to true the watch returns the first change under the given
// prefix since the given index.
//
// If recursive is set to false the watch returns the first change to the given key
// since the given index.
//
// To watch for the latest change, set waitIndex = 0.
//
// If a receiver channel is given, it will be a long-term watch. Watch will block at the
//channel. After someone receives the channel, it will go on to watch that
// prefix.  If a stop channel is given, the client can close long-term watch using
// the stop channel.
func (c *Client) Watch(prefix string, waitIndex uint64, recursive bool,
	receiver chan *Response, stop chan bool) (*Response, error) {
	logger.Debugf("watch %s [%s]", prefix, c.cluster.Leader)
	if receiver == nil {
		raw, err := c.watchOnce(prefix, waitIndex, recursive, stop)

		if err != nil {
			return nil, err
		}

		return raw.toResponse()
	}

	for {
		raw, err := c.watchOnce(prefix, waitIndex, recursive, stop)

		if err != nil {
			return nil, err
		}

		resp, err := raw.toResponse()

		if err != nil {
			return nil, err
		}

		waitIndex = resp.Node.ModifiedIndex + 1
		receiver <- resp
	}

	return nil, nil
}

func (c *Client) RawWatch(prefix string, waitIndex uint64, recursive bool,
	receiver chan *RawResponse, stop chan bool) (*RawResponse, error) {

	logger.Debugf("rawWatch %s [%s]", prefix, c.cluster.Leader)
	if receiver == nil {
		return c.watchOnce(prefix, waitIndex, recursive, stop)
	}

	for {
		raw, err := c.watchOnce(prefix, waitIndex, recursive, stop)

		if err != nil {
			return nil, err
		}

		resp, err := raw.toResponse()

		if err != nil {
			return nil, err
		}

		waitIndex = resp.Node.ModifiedIndex + 1
		receiver <- raw
	}

	return nil, nil
}

func (c *Client) AsyncWatch(prefix string, waitIndex uint64, recursive bool) (<-chan *Response, chan<- bool, <-chan error, error) {
	logger.Debugf("watch %s [%s]", prefix, c.cluster.Leader)

	options := options{
		"wait": true,
	}
	if waitIndex > 0 {
		options["waitIndex"] = waitIndex
	}
	if recursive {
		options["recursive"] = true
	}

	innerRespChan, stopChan, innerErrChan, err := c.asyncGet(prefix, options)

	if err != nil {
		return nil, nil, nil, err
	}

	respChan := make(chan *Response)
	errChan := make(chan error)

	go func() {
		defer close(respChan)
		defer close(errChan)

		select {
		case raw, ok := <-innerRespChan:
			if !ok {
				return
			}
			response, err := raw.toResponse()
			if err != nil {
				errChan <- err
				return
			}
			respChan <- response
		case err, ok := <-innerErrChan:
			if !ok {
				return
			}

			errChan <- err
		}
	}()

	return respChan, stopChan, errChan, nil
}

// helper func
// return when there is change under the given prefix
func (c *Client) watchOnce(key string, waitIndex uint64, recursive bool, stop chan bool) (*RawResponse, error) {
	respChan := make(chan *RawResponse, 1)
	errChan := make(chan error)

	go func() {
		options := options{
			"wait": true,
		}
		if waitIndex > 0 {
			options["waitIndex"] = waitIndex
		}
		if recursive {
			options["recursive"] = true
		}

		resp, err := c.get(key, options)

		if err != nil {
			errChan <- err
			return
		}

		respChan <- resp
	}()

	select {
	case resp := <-respChan:
		return resp, nil
	case err := <-errChan:
		return nil, err
	case <-stop:
		return nil, ErrWatchStoppedByUser
	}
}
