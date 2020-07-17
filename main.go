package main

import (
	"context"
	"github.com/nnqq/scr-url-producer/config"
	"github.com/nnqq/scr-url-producer/producer"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, os.Kill)
	go func() {
		<-exit
		cancel()
	}()

	producer.URL(ctx, config.Env.FilePath)
}
