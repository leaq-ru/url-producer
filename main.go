package main

import (
	graceful "github.com/nnqq/scr-lib-graceful"
	"github.com/nnqq/scr-url-producer/logger"
	"github.com/nnqq/scr-url-producer/url"
	"net/http"
)

func healthz() {
	http.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write(nil)
		logger.Err(err)
	}))
	logger.Err(http.ListenAndServe("0.0.0.0:80", nil))
}

func main() {
	healthz()

	urlProducer := url.NewProducer()
	go graceful.HandleSignals(urlProducer.GracefulStop)

	logger.Must(urlProducer.Run())
}
