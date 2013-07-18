package main 

import (
	"github.com/coreos/go-etcd/etcd"
	"fmt"
	"time"
)

var count = 0

func main() {
	etcd.SyncCluster()
	c := make(chan bool, 10)
	// set up a lock
	for i:=0; i < 200; i++ {
		go t(i, c)
	}
	start := time.Now()
	for i:=0; i< 200; i++ {
		<-c
	}
	fmt.Println(time.Now().Sub(start), ": ", 200 * 100, "commands")
}

func t(num int,c chan bool) {
	for i := 0; i < 100; i++ {
		str := fmt.Sprintf("foo_%d",num * i)
		etcd.Set(str, "10", 0)
	}
	c<-true
}



