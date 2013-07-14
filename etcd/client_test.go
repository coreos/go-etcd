package etcd

import (
	"fmt"
	"testing"
)

// To pass this test, we need to create a cluster of 3 machines
// The server should be listening on 127.0.0.1:4001, 4002, 4003
func TestSync(t *testing.T) {
	success := syncMachines()
	if !success {
		t.Fatal("cannot sync machines")
	} else {
		fmt.Println(client.cluster.Machines)
	}
}
