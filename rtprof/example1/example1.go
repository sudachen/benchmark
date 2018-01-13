package main

import (
	"os"
	"os/signal"
	"time"

	"github.com/sudachen/benchmark/rtprof"
)

func main() {
	rtppf.Start(5*time.Second, 8080)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	signal.Stop(c)
	rtppf.Stop()
}
