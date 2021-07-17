package main

import (
	graceful "github.com/leaq-ru/lib-graceful"
	"github.com/leaq-ru/url-producer/logger"
	"github.com/leaq-ru/url-producer/url"
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
	go healthz()

	urlProducer := url.NewProducer()
	go graceful.HandleSignals(urlProducer.GracefulStop)

	logger.Must(urlProducer.Run())
}
