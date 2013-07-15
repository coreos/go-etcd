package main 

import (
	"github.com/coreos/go-etcd/etcd"
	"fmt"
)

var count = 0

func main() {
	etcd.SyncCluster()
	c := make(chan bool, 10)
	// set up a lock
	etcd.Set("lock", "unlock", 0)
	for i:=0; i < 10; i++ {
		go t(i, c)
	}

	for i:=0; i< 10; i++ {
		<-c
	}
}

func t(num int,c chan bool) {
	for i := 0; i < 100; i++ {
		lock()
		// a stupid spin lock
		count++
		fmt.Println(num, " got the lock and update count to", count)
		unlock()
		fmt.Println(num, " released the lock")
	}
	c<-true
}


// A stupid spin lock
func lock() {
	for {
		_, success, _ := etcd.TestAndSet("lock", "unlock", "lock", 0)
		if success != true {
			fmt.Println("tried lock failed!")
		} else {
			return
		}
	}
}

func unlock() {
	for {
		_, err := etcd.Set("lock", "unlock", 0)
		if err == nil{
			return
		}
		fmt.Println(err)
	}
}


