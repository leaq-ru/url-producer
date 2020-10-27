package url

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"github.com/nnqq/scr-url-producer/config"
	"github.com/nnqq/scr-url-producer/logger"
	"github.com/nnqq/scr-url-producer/mongo"
	"github.com/nnqq/scr-url-producer/protocol"
	"github.com/nnqq/scr-url-producer/stan"
	"go.mongodb.org/mongo-driver/bson"
	m "go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type producer struct {
	done chan struct{}
}

type fileEntity struct {
	Row string `bson:"r"`
}

func NewProducer() *producer {
	return &producer{
		done: make(chan struct{}, 1),
	}
}

func (p *producer) Run() (err error) {
	count, err := mongo.FileEntity.CountDocuments(context.Background(), bson.D{})
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return
	}

	if count == 0 {
		var g errgroup.Group
		g.Go(func() error {
			return loadListToMongo(config.Env.DomainsFile.URL)
		})
		g.Go(func() error {
			return loadListToMongo(config.Env.DomainsFile.URLsu)
		})
		g.Go(func() error {
			return loadListToMongo(config.Env.DomainsFile.URLrf)
		})
		err = g.Wait()
		if err != nil {
			logger.Log.Error().Err(err).Send()
			return
		}
	}

	err = p.startMongoProcessing()
	if err != nil {
		logger.Log.Error().Err(err).Send()
	}
	return
}

func loadListToMongo(url string) (err error) {
	res, err := http.Get(url)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return
	}
	err = res.Body.Close()
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(body))
	body = nil
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return
	}

	fileBytes, err := ioutil.ReadAll(gzipReader)
	gzipReader = nil
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return
	}

	file := strings.Split(string(fileBytes), "\n")
	fileBytes = nil

	for _, row := range file {
		_, err = mongo.FileEntity.InsertOne(context.Background(), fileEntity{
			Row: row,
		})
		if err != nil {
			logger.Log.Error().Err(err).Send()
			return
		}
	}
	file = nil
	return
}

func (p *producer) startMongoProcessing() (err error) {
	for {
		select {
		case <-p.done:
			logger.Log.Debug().Msg("URL producer loop graceful shutdown")
			return
		default:
			err = processRow()
			if err != nil {
				if errors.Is(err, m.ErrNoDocuments) {
					return nil
				}

				logger.Log.Error().Err(err).Send()
				return
			}
		}
	}
}

func (p *producer) GracefulStop() {
	close(p.done)
}

// format
// SOMETHING.RU	REGRU-RU	18.11.2015	18.11.2020	19.12.2020	1
func sendLine(line string) (err error) {
	values := strings.Split(line, "\t")

	url := strings.ToLower(values[0])
	registar := strings.ToLower(values[1])
	registrationDate, err := time.Parse("02.01.2006", values[2])
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return
	}

	b, err := json.Marshal(protocol.URLMessage{
		URL:              url,
		Registrar:        registar,
		RegistrationDate: registrationDate,
	})
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return
	}

	err = stan.Conn.Publish("url", b)
	if err != nil {
		logger.Log.Error().Err(err).Send()
	}
	return
}

func processRow() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	fe := fileEntity{}
	err = mongo.FileEntity.FindOneAndDelete(ctx, bson.D{}).Decode(&fe)
	if err != nil {
		return
	}

	if fe.Row == "" {
		return
	}

	err = sendLine(fe.Row)
	if err != nil {
		logger.Log.Error().Err(err).Send()
	}
	return
}
