package main

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
)

var count = 0

func main() {

	good := 0
	bad := 0


	// open https
	// etcd.SetScheme(etcd.HTTPS)

	// add client cert
	// success, err :=etcd.SetCertAndKey("key/me_cert.pem", "key/me_key.pem")

	// if !success {
	// 	fmt.Println(err)
	// }

	etcd.SyncCluster()
	c := make(chan bool, 10)
	// set up a lock
	etcd.Set("lock", "unlock", 0)
	for i := 0; i < 10; i++ {
		go t(i, c)
	}

	for i := 0; i < 10; i++ {
		if <-c {
			good++
		} else {
			bad++
		}
	}
	fmt.Println("good: ", good, "bad: ", bad)
}

func t(num int, c chan bool) {
	for i := 0; i < 100; i++ {
		if lock() {
			// a stupid spin lock
			count++
			fmt.Println(num, " got the lock and update count to", count)
			unlock()
			fmt.Println(num, " released the lock")
		} else {
			c <- false
			return
		}
	}
	c <- true
}

// A stupid spin lock
func lock() bool {
	for {
		_, success, err := etcd.TestAndSet("lock", "unlock", "lock", 0)

		if err != nil {
			return false
		}

		if success != true {
			fmt.Println("tried lock failed!")
		} else {
			return true
		}
	}
}

func unlock() {
	for {
		_, err := etcd.Set("lock", "unlock", 0)
		if err == nil {
			return
		}
		fmt.Println(err)
	}
}
