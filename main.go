package main

import (
	"github.com/nnqq/scr-url-producer/config"
	"github.com/nnqq/scr-url-producer/logger"
	"github.com/nnqq/scr-url-producer/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func handleSignals(stopFunc ...func() error) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	<-signals
	wg := sync.WaitGroup{}
	wg.Add(len(stopFunc))
	for _, f := range stopFunc {
		logger.Err(f())
	}
	wg.Wait()
}

func main() {
	urlProducer := url.NewProducer()
	go handleSignals(urlProducer.GracefulStop)

	logger.Err(urlProducer.Run(config.Env.FilePath))
}
