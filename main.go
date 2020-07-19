package main

import (
	"context"
	"github.com/nnqq/scr-url-producer/config"
	"github.com/nnqq/scr-url-producer/logger"
	"github.com/nnqq/scr-url-producer/mongo"
	"github.com/nnqq/scr-url-producer/producer"
	"github.com/nnqq/scr-url-producer/stan"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func cleanup(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	e := func(err error) {
		if err != nil {
			logger.Log.Error().Err(err).Send()
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		e(stan.Conn.Close())
	}()

	go func() {
		defer wg.Done()
		e(mongo.DB.Client().Disconnect(ctx))
	}()
	wg.Wait()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signals
		cleanup(ctx)
		cancel()
	}()

	producer.URL(ctx, config.Env.FilePath)
}
