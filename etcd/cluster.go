package etcd

import (
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"sync"
)

type Cluster struct {
	Leader          string   `json:"leader"`
	Machines        []string `json:"machines"`
	healthyMachines []string
	//picked   int
	picked string
	mu     sync.RWMutex
}

func NewCluster(machines []string) *Cluster {
	// if an empty slice was sent in then just assume HTTP 4001 on localhost
	if len(machines) == 0 {
		machines = []string{"http://127.0.0.1:4001"}
	}

	machines = shuffleStringSlice(machines)
	logger.Debug("Shuffle cluster machines", machines)
	// default leader and machines
	return &Cluster{
		Leader:          "",
		Machines:        machines,
		healthyMachines: machines,
		picked:          machines[rand.Intn(len(machines))],
	}
}

func (cl *Cluster) failure() {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.picked = cl.healthyMachines[rand.Intn(len(cl.healthyMachines))]
}

func (cl *Cluster) pick() string {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	return cl.picked
}

func (cl *Cluster) updateFromStr(machines string) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.Machines = strings.Split(machines, ",")
	for i := range cl.Machines {
		cl.Machines[i] = strings.TrimSpace(cl.Machines[i])
	}

	cl.healthyMachines = shuffleStringSlice(cl.Machines)
	for _, healthyMachine := range cl.healthyMachines {
		if healthyMachine == cl.picked {
			return
		}
	}
	cl.picked = cl.healthyMachines[rand.Intn(len(cl.healthyMachines))]
}

func (cl *Cluster) updateHealthyMachines(newHealthyMachines []string) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	sort.Sort(sort.StringSlice(newHealthyMachines))

	oldHealthyMachines := cl.healthyMachines
	sort.Sort(sort.StringSlice(oldHealthyMachines))

	if reflect.DeepEqual(oldHealthyMachines, newHealthyMachines) {
		return
	}

	cl.healthyMachines = newHealthyMachines
	cl.healthyMachines = shuffleStringSlice(cl.healthyMachines)
}
