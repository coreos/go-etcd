package etcd

import (
	"github.com/coreos/go-log/log"
	"os"
)

var logger *log.Logger

func init() {
	setLogger(log.PriErr)
	// OpenDebug()
}

func OpenDebug() {
	setLogger(log.PriDebug)
}

func CloseDebug() {
	setLogger(log.PriErr)
}

func setLogger(priority log.Priority) {
	logger = log.NewSimple(
		log.PriorityFilter(
			priority,
			log.WriterSink(os.Stdout, log.BasicFormat, log.BasicFields)))
}
