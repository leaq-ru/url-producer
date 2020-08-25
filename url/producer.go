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
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type producer struct {
	done chan struct{}
}

type fileOffset struct {
	Index int `bson:"index"`
}

func NewProducer() *producer {
	return &producer{
		done: make(chan struct{}, 1),
	}
}

func (p *producer) Run() (err error) {
	res, err := http.Get(config.Env.DomainsFile.URL)
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
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return
	}

	fileBytes, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return
	}

	file := strings.Split(string(fileBytes), "\n")
	fileLastIndex := len(file) - 1

	for {
		select {
		case <-p.done:
			logger.Log.Debug().Msg("URL producer loop graceful shutdown")
			return
		default:
			err = processRow(file, fileLastIndex)
			if err != nil {
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

func processRow(file []string, fileLastIndex int) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	offset := fileOffset{}
	err = mongo.FileOffset.FindOne(ctx, bson.D{}).Decode(&offset)
	if err != nil && !errors.Is(err, m.ErrNoDocuments) {
		logger.Log.Error().Err(err).Send()
		return
	}

	if offset.Index == fileLastIndex {
		_, err = mongo.FileOffset.DeleteOne(ctx, bson.D{})
		if err != nil {
			logger.Log.Error().Err(err).Send()
			return
		}

		logger.Log.Debug().Msg("file iteration done")
		return
	}

	err = sendLine(file[offset.Index])
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return
	}
	logger.Log.Debug().Str("domain", file[offset.Index]).Msg("sent URL via NATS streaming")

	opts := options.Update()
	opts.SetUpsert(true)
	_, err = mongo.FileOffset.UpdateOne(ctx, bson.D{}, bson.M{
		"$inc": bson.M{
			"index": 1,
		},
	}, opts)
	if err != nil {
		logger.Log.Error().Err(err).Send()
	}
	return
}
