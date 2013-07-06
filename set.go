package goetcd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/xiangli-cmu/raft-etcd/store"
	"io/ioutil"
	"net/http"
	"path"
)

var version = "v1"

func Set(cluster string, key string, value string, ttl uint64) (*store.Response, error) {

	httpPath := path.Join(cluster, "/", version, "/keys/", key)

	//TODO: deal with https
	httpPath = "http://" + httpPath

	content := "value=" + value

	if ttl > 0 {
		content += fmt.Sprintf("&ttl=%v", ttl)
	}

	reader := bytes.NewReader([]byte(content))

	resp, err := http.Post(httpPath, "application/x-www-form-urlencoded", reader)

	if resp != nil {
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return nil, err
		}

		var result store.Response

		err = json.Unmarshal(b, &result)

		if err != nil {
			return nil, err
		}

		return &result, nil

	}

	return nil, err

}
