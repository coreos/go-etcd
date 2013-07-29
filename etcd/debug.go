package etcd

import (
	"github.com/ccding/go-logging/logging"
)

var logger = logging.SimpleLogger("go-etcd")

func init() {
	logger.SetLevel(logging.NOTSET)
}
