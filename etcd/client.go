package etcd

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"path"
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
		Scheme: "http://",
		// default timeout is one second
		Timeout: time.Second,
	}

	tr := &http.Transport{
		Dial: dialTimeout,
	}

	client = Client{
		cluster:    cluster,
		config:     config,
		httpClient: &http.Client{Transport: tr},
	}

}

func SetCert(cert string, key string) (bool, error) {

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
	}
	return false, errors.New("Require both cert and key path")
}

func SetScheme(scheme int) (bool, error) {
	if scheme == HTTP {
		client.config.Scheme = "http://"
		return true, nil
	}
	if scheme == HTTPS {
		client.config.Scheme = "https://"
		return true, nil
	}
	return false, errors.New("Unknow Scheme")
}

// Try to connect from the given machine
// Store the cluster information in the given machineFile
func setCluster(machine string, machineFile string, confFile string) {

}

// Try to connect from the given machines in the file
// Update the cluster information in the given machineFile
func setClusterFromFile(machineFile string, confFile string) {

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

	httpPath = client.config.Scheme + httpPath
	return httpPath
}
