package gworker

import (
	"os"
	"os/signal"
	"syscall"
)

func HandleSignal() chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)
	return c
}
