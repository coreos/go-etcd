package etcd

import (
	"github.com/ccding/go-logging/logging"
)

var logger, _ = logging.SimpleLogger("go-etcd")

func init() {
	logger.SetLevel(logging.ERROR)
}
