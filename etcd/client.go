package etcd

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"path"
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
	cluster    Cluster
	config     Config
	httpClient *http.Client
}

var client Client

// Setup a basic conf and cluster
func init() {

	// default leader and machines
	cluster := Cluster{
		Leader:   "0.0.0.0:4001",
		Machines: make([]string, 1),
	}
	cluster.Machines[0] = "0.0.0.0:4001"

	config := Config{
		// default use http
		Scheme: "http",
		// default timeout is one second
		Timeout: time.Second,
	}

	tr := &http.Transport{
		Dial: dialTimeout,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client = Client{
		cluster:    cluster,
		config:     config,
		httpClient: &http.Client{Transport: tr},
	}

}

func SetCertAndKey(cert string, key string) (bool, error) {

	if cert != "" && key != "" {
		tlsCert, err := tls.LoadX509KeyPair(cert, key)

		if err != nil {
			return false, err
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{tlsCert},
				InsecureSkipVerify: true,
			},
			Dial: dialTimeout,
		}

		client.httpClient = &http.Client{Transport: tr}
		return true, nil
	}
	return false, errors.New("Require both cert and key path")
}

func SetScheme(scheme int) (bool, error) {
	if scheme == HTTP {
		client.config.Scheme = "http"
		return true, nil
	}
	if scheme == HTTPS {
		client.config.Scheme = "https"
		return true, nil
	}
	return false, errors.New("Unknow Scheme")
}

// Try to sync from the given machine
func SetCluster(machines []string) bool {
	success := internalSyncCluster(machines)
	return success
}

// Try to connect from the given machines in the file
func setClusterFromFile(machineFile string, confFile string) {

}

// sycn cluster information using the existing machine list
func SyncCluster() bool {
	success := internalSyncCluster(client.cluster.Machines)
	return success
}

// sync cluster information by providing machine list
func internalSyncCluster(machines []string) bool {
	for _, machine := range machines {
		httpPath := createHttpPath(machine, "machines")
		resp, err := client.httpClient.Get(httpPath)
		if err != nil {
			// try another machine in the cluster
			continue
		} else {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				// try another machine in the cluster
				continue
			}
			// update Machines List
			client.cluster.Machines = strings.Split(string(b), ",")
			return true
		}
	}
	return false
}

// serverName should contain both hostName and port
func createHttpPath(serverName string, _path string) string {
	httpPath := path.Join(serverName, _path)
	httpPath = client.config.Scheme + "://" + httpPath
	return httpPath
}

// Dial with timeout
func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, (client.config.Timeout))
}

func getHttpPath(s ...string) string {
	httpPath := path.Join(client.cluster.Leader, version)

	for _, seg := range s {
		httpPath = path.Join(httpPath, seg)
	}

	httpPath = client.config.Scheme + "://" + httpPath
	return httpPath
}

func updateLeader(httpPath string) {
	// httpPath http://127.0.0.1:4001/v1...
	leader := strings.Split(httpPath, "://")[1]
	// we want to have 127.0.0.1:4001

	leader = strings.Split(httpPath, "/")[0]
	client.cluster.Leader = leader
}

// Wrap GET, POST and internal error handling
func sendRequest(method string, _path string, body string) (*http.Response, error) {

	var resp *http.Response
	var err error
	var req *http.Request

	retry := 0
	// if we connect to a follower, we will retry until we found a leader
	for {

		httpPath := getHttpPath(_path)

		if body == "" {

			req, _ = http.NewRequest(method, httpPath, nil)

		} else {
			req, _ = http.NewRequest(method, httpPath, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
		}

		resp, err = client.httpClient.Do(req)

		// network error, change a machine!
		if err != nil {
			retry++
			if retry > 2*len(client.cluster.Machines) {
				return nil, err
			}
			num := retry % len(client.cluster.Machines)
			client.cluster.Leader = client.cluster.Machines[num]
			continue
		}

		if resp != nil {
			if resp.StatusCode == http.StatusTemporaryRedirect {
				httpPath := resp.Header.Get("Location")

				resp.Body.Close()

				if httpPath == "" {
					return nil, errors.New("Cannot get redirection location")
				}

				updateLeader(httpPath)

				// try to connect the leader
				continue
			} else {
				break
			}

		}
		return nil, err
	}
	return resp, nil
}
