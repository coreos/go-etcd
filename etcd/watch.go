package etcd

import (
	"errors"
)

// Errors introduced by the Watch command.
var (
	ErrWatchStoppedByUser = errors.New("Watch stopped by the user via stop channel")
)

// If recursive set to true, watch returns the first change under the given
// prefix since the given index.

// If recursive set to false, watch returns the first change to the given key
// since the given index.
// To watch for the latest change, set waitIndex = 0.
//
// If a receiver channel is given, it will be a long-term watch. Watch will block at the
// channel. And after someone receive the channel, it will go on to watch that
// prefix.  If a stop channel is given, client can close long-term watch using
// the stop channel
func (c *Client) Watch(prefix string, waitIndex uint64, recursive bool,
	receiver chan *Response, stop chan bool) (*Response, error) {

	logger.Debugf("watch %s [%s]", prefix, c.cluster.Leader)
	if receiver == nil {
		raw, err := c.watchOnce(prefix, waitIndex, recursive, stop)

		if err != nil {
			return nil, err
		}

		return raw.toResponse()
	} else {
		for {
			raw, err := c.watchOnce(prefix, waitIndex, recursive, stop)

			if err != nil {
				return nil, err
			}

			resp, err := raw.toResponse()

			if resp != nil {
				waitIndex = resp.ModifiedIndex + 1
				receiver <- resp
			} else {
				return nil, err
			}
		}
	}

	return nil, nil
}

func (c *Client) RawWatch(prefix string, waitIndex uint64, recursive bool,
	receiver chan *RawResponse, stop chan bool) (*RawResponse, error) {

	logger.Debugf("rawWatch %s [%s]", prefix, c.cluster.Leader)
	if receiver == nil {
		return c.watchOnce(prefix, waitIndex, recursive, stop)
	} else {
		for {
			raw, err := c.watchOnce(prefix, waitIndex, recursive, stop)

			if err != nil {
				return nil, err
			}

			resp, err := raw.toResponse()

			if resp != nil {
				waitIndex = resp.ModifiedIndex + 1
				receiver <- raw
			} else {
				return nil, err
			}
		}
	}

	return nil, nil
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
