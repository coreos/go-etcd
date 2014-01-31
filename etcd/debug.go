package etcd

import (
	"io/ioutil"
	"log"
	"strings"
)

var logger log.Logger

type defaultLogger struct {
	log *log.Logger
}

func SetLogger(l defaultLogger) {
	logger = *l.log
}

func GetLogger() log.Logger {
	return logger
}

func (p *defaultLogger) Debug(args ...interface{}) {
	p.log.Println(args)
}

func (p *defaultLogger) Debugf(fmt string, args ...interface{}) {
	// Append newline if necessary
	if !strings.HasSuffix(fmt, "\n") {
		fmt = fmt + "\n"
	}
	p.log.Printf(fmt, args)
}

func (p *defaultLogger) Warning(args ...interface{}) {
	p.Debug(args)
}

func (p *defaultLogger) Warningf(fmt string, args ...interface{}) {
	p.Debugf(fmt, args)
}

func init() {
	// Default logger uses the go default log.
	SetLogger(defaultLogger{log.New(ioutil.Discard, "go-etcd", log.LstdFlags)})
}
