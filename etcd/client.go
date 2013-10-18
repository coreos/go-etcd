package etcd

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"
	"time"
)

const (
	HTTP = iota
	HTTPS
)

type Cluster struct {
	Leader   string   `json:"leader"`
	Machines []string `json:"machines"`
}

type Config struct {
	CertFile string        `json:"certFile"`
	KeyFile  string        `json:"keyFile"`
	Scheme   string        `json:"scheme"`
	Timeout  time.Duration `json:"timeout"`
}

type Client struct {
	cluster     Cluster `json:"cluster"`
	config      Config  `json:"config"`
	httpClient  *http.Client
	persistence io.Writer
}

type Options map[string]interface{}

// An internally-used data structure that represents a mapping
// between valid options and their kinds
type validOptions map[string]reflect.Kind

// NewClient create a basic client that is configured to be used
// with the given machine list.
func NewClient(machines []string) *Client {
	// if an empty slice was sent in then just assume localhost
	if len(machines) == 0 {
		machines = []string{"http://127.0.0.1:4001"}
	}

	// default leader and machines
	cluster := Cluster{
		Leader:   machines[0],
		Machines: machines,
	}

	config := Config{
		// default use http
		Scheme: "http",
		// default timeout is one second
		Timeout: time.Second,
	}

	client := &Client{
		cluster: cluster,
		config:  config,
	}

	err := setupHttpClient(client)
	if err != nil {
		panic(err)
	}

	return client
}

// NewClientFile creates a client from a given file path.
// The given file is expected to use the JSON format.
func NewClientFile(fpath string) (*Client, error) {
	fi, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	return NewClientReader(fi)
}

// NewClientReader creates a Client configured from a given reader.
// The config is expected to use the JSON format.
func NewClientReader(reader io.Reader) (*Client, error) {
	var client Client

	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &client)
	if err != nil {
		return nil, err
	}

	err = setupHttpClient(&client)
	if err != nil {
		return nil, err
	}

	return &client, nil
}

func setupHttpClient(client *Client) error {
	if client.config.CertFile != "" && client.config.KeyFile != "" {
		err := client.SetCertAndKey(client.config.CertFile, client.config.KeyFile)
		if err != nil {
			return err
		}
	} else {
		client.config.CertFile = ""
		client.config.KeyFile = ""
		tr := &http.Transport{
			Dial: dialTimeout,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		client.httpClient = &http.Client{Transport: tr}
	}

	return nil
}

// SetPersistence sets a writer to which the config will be
// written every time it's changed.
func (c *Client) SetPersistence(writer io.Writer) {
	c.persistence = writer
}

// MarshalJSON implements the Marshaller interface
// as defined by the standard JSON package.
func (c *Client) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(struct {
		Config  Config  `json:"config"`
		Cluster Cluster `json:"cluster"`
	}{
		Config:  c.config,
		Cluster: c.cluster,
	})

	if err != nil {
		return nil, err
	}

	return b, nil
}

// UnmarshalJSON implements the Unmarshaller interface
// as defined by the standard JSON package.
func (c *Client) UnmarshalJSON(b []byte) error {
	temp := struct {
		Config  Config  `json: "config"`
		Cluster Cluster `json: "cluster"`
	}{}
	err := json.Unmarshal(b, &temp)
	if err != nil {
		return err
	}

	c.cluster = temp.Cluster
	c.config = temp.Config
	return nil
}

// saveConfig saves the current config using c.persistence.
func (c *Client) saveConfig() error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	_, err = c.persistence.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) SetCertAndKey(cert string, key string) error {
	if cert != "" && key != "" {
		tlsCert, err := tls.LoadX509KeyPair(cert, key)

		if err != nil {
			return err
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{tlsCert},
				InsecureSkipVerify: true,
			},
			Dial: dialTimeout,
		}

		c.httpClient = &http.Client{Transport: tr}
		return nil
	}
	return errors.New("Require both cert and key path")
}

func (c *Client) SetScheme(scheme int) (bool, error) {
	if scheme == HTTP {
		c.config.Scheme = "http"
		return true, nil
	}
	if scheme == HTTPS {
		c.config.Scheme = "https"
		return true, nil
	}
	return false, errors.New("Unknown Scheme")
}

// SetCluster updates config using the given machine list.
func (c *Client) SetCluster(machines []string) bool {
	success := c.internalSyncCluster(machines)
	return success
}

func (c *Client) GetCluster() []string {
	return c.cluster.Machines
}

// SyncCluster updates config using the internal machine list.
func (c *Client) SyncCluster() bool {
	success := c.internalSyncCluster(c.cluster.Machines)
	return success
}

// internalSyncCluster syncs cluster information using the given machine list.
func (c *Client) internalSyncCluster(machines []string) bool {
	for _, machine := range machines {
		httpPath := c.createHttpPath(machine, version+"/machines")
		resp, err := c.httpClient.Get(httpPath)
		if err != nil {
			// try another machine in the cluster
			continue
		} else {
			b, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				// try another machine in the cluster
				continue
			}

			// update Machines List
			c.cluster.Machines = strings.Split(string(b), ", ")

			// update leader
			// the first one in the machine list is the leader
			logger.Debugf("update.leader[%s,%s]", c.cluster.Leader, c.cluster.Machines[0])
			c.cluster.Leader = c.cluster.Machines[0]

			logger.Debug("sync.machines ", c.cluster.Machines)
			return true
		}
	}
	return false
}

// createHttpPath creates a complete HTTP URL.
// serverName should contain both the host name and a port number, if any.
func (c *Client) createHttpPath(serverName string, _path string) string {
	u, _ := url.Parse(serverName)
	u.Path = path.Join(u.Path, "/", _path)

	if u.Scheme == "" {
		u.Scheme = "http"
	}
	return u.String()
}

// Dial with timeout.
func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, time.Second)
}

func (c *Client) getHttpPath(s ...string) string {
	fullPath := c.cluster.Leader + "/" + version
	for _, seg := range s {
		fullPath = fullPath + "/" + seg
	}

	return fullPath
}

func (c *Client) updateLeader(httpPath string) {
	u, _ := url.Parse(httpPath)

	var leader string
	if u.Scheme == "" {
		leader = "http://" + u.Host
	} else {
		leader = u.Scheme + "://" + u.Host
	}

	logger.Debugf("update.leader[%s,%s]", c.cluster.Leader, leader)
	c.cluster.Leader = leader
}

// Wrap GET, POST and internal error handling
func (c *Client) sendRequest(method string, _path string, body string) (*http.Response, error) {

	var resp *http.Response
	var req *http.Request

	retry := 0
	// if we connect to a follower, we will retry until we found a leader
	for {
		var httpPath string

		// If _path has schema already, then it's assumed to be
		// a complete URL and therefore needs no further processing.
		u, err := url.Parse(_path)
		if err != nil {
			return nil, err
		}

		if u.Scheme != "" {
			httpPath = _path
		} else {
			httpPath = c.getHttpPath(_path)
		}

		logger.Debug("send.request.to ", httpPath, " | method ", method)
		if body == "" {

			req, _ = http.NewRequest(method, httpPath, nil)

		} else {
			req, _ = http.NewRequest(method, httpPath, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		}

		resp, err = c.httpClient.Do(req)

		logger.Debug("recv.response.from ", httpPath)
		// network error, change a machine!
		if err != nil {
			retry++
			if retry > 2*len(c.cluster.Machines) {
				return nil, errors.New("Cannot reach servers")
			}
			num := retry % len(c.cluster.Machines)
			logger.Debug("update.leader[", c.cluster.Leader, ",", c.cluster.Machines[num], "]")
			c.cluster.Leader = c.cluster.Machines[num]
			time.Sleep(time.Millisecond * 200)
			continue
		}

		if resp != nil {
			if resp.StatusCode == http.StatusTemporaryRedirect {
				httpPath := resp.Header.Get("Location")

				resp.Body.Close()

				if httpPath == "" {
					return nil, errors.New("Cannot get redirection location")
				}

				c.updateLeader(httpPath)
				logger.Debug("send.redirect")
				// try to connect the leader
				continue
			} else if resp.StatusCode == http.StatusInternalServerError {
				resp.Body.Close()

				retry++
				if retry > 2*len(c.cluster.Machines) {
					return nil, errors.New("Cannot reach servers")
				}
				continue
			} else {
				logger.Debug("send.return.response ", httpPath)
				break
			}

		}
		logger.Debug("error.from ", httpPath, " ", err.Error())
		return nil, err
	}
	return resp, nil
}
