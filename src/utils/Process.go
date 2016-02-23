package utils

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
)

// Stop empty struct used send a signal to a process thank to a channel.
type Stop struct{}

// ProcessInterface interface for process
type ProcessInterface interface {
	Init() error
	Clear()
	Start(stop chan Stop) error
}

// ExecProcess execute a process
func ExecProcess(p ProcessInterface) error {
	p.Init()
	defer p.Clear()

	glog.Info("Starting...")
	sigs := make(chan os.Signal, 1)
	stop := make(chan Stop, 1)

	signal.Notify(sigs, syscall.SIGINT)
	signal.Notify(sigs, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		glog.Info("Signal: ", sig)
		stop <- Stop{}
	}()

	return p.Start(stop)
}
