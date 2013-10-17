package etcd

import (
	"crypto/tls"
	"errors"
	"github.com/BurntSushi/toml"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
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
	Leader   string
	Machines []string
}

type Config struct {
	CertFile string
	KeyFile  string
	Scheme   string
	Timeout  time.Duration
}

type Client struct {
	cluster     Cluster
	config      Config
	httpClient  *http.Client
	persistence io.Writer
}

type Options map[string]interface{}

// An internally-used data structure that represents a mapping
// between valid options and their kinds
type validOptions map[string]reflect.Kind

// Setup a basic conf and cluster
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

	// Get the list of currently running nodes
	client.SyncCluster()
	return client
}

// Construct a client from a given config file in TOML format
func NewClientFile(fpath string) (*Client, error) {
	var client *Client
	if _, err := toml.DecodeFile(fpath, client); err != nil {
		return nil, err
	}

	err := setupHttpClient(client)
	if err != nil {
		return nil, err
	}

	client.SyncCluster()
	return client, nil
}

// Construct a client by reading a config file in TOML format
// from the given reader
func NewClientReader(reader io.Reader) (*Client, error) {
	var client *Client
	if _, err := toml.DecodeReader(reader, client); err != nil {
		return nil, err
	}

	err := setupHttpClient(client)
	if err != nil {
		return nil, err
	}

	client.SyncCluster()
	return client, nil
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

// Set a writer to which the config will be written every time it's changed.
func (c *Client) SetPersistence(writer io.Writer) {
	c.persistence = writer
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

// Try to sync from the given machine
func (c *Client) SetCluster(machines []string) bool {
	success := c.internalSyncCluster(machines)
	return success
}

func (c *Client) GetCluster() []string {
	return c.cluster.Machines
}

// sycn cluster information using the existing machine list
func (c *Client) SyncCluster() bool {
	success := c.internalSyncCluster(c.cluster.Machines)
	return success
}

// sync cluster information by providing machine list
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

func (c *Client) saveConfig() error {
	// TODO
	return nil
}

// serverName should contain both hostName and port
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
	var err error
	var req *http.Request

	retry := 0
	// if we connect to a follower, we will retry until we found a leader
	for {
		var httpPath string
		if strings.HasPrefix(_path, "http") {
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
