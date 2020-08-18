package main

import (
	graceful "github.com/nnqq/scr-lib-graceful"
	"github.com/nnqq/scr-url-producer/config"
	"github.com/nnqq/scr-url-producer/logger"
	"github.com/nnqq/scr-url-producer/url"
)

func main() {
	urlProducer := url.NewProducer()
	go graceful.HandleSignals(urlProducer.GracefulStop)

	logger.Must(urlProducer.Run(config.Env.DomainsFileURL))
}
