package producer

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/nnqq/scr-url-producer/logger"
	"github.com/nnqq/scr-url-producer/mongo"
	"github.com/nnqq/scr-url-producer/protocol"
	"github.com/nnqq/scr-url-producer/stan"
	"go.mongodb.org/mongo-driver/bson"
	m "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"strings"
	"time"
)

type fileOffset struct {
	Index int `bson:"index"`
}

func URL(ctx context.Context, localPath string) {
	fileBytes, err := ioutil.ReadFile(localPath)
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return
	}

	file := strings.Split(string(fileBytes), "\n")
	fileLastIndex := len(file) - 1

	for {
		select {
		case <-ctx.Done():
			logger.Log.Debug().Msg("URL producer loop exit")
			return
		default:
			offset := fileOffset{}
			err := mongo.FileOffset.FindOne(ctx, bson.D{}).Decode(&offset)
			if err != nil && !errors.Is(err, m.ErrNoDocuments) {
				logger.Log.Error().Err(err).Send()
				return
			}

			if offset.Index == fileLastIndex {
				logger.Log.Debug().Msg("file iteration done")
				return
			}

			err = sendLine(file[offset.Index])
			if err != nil {
				logger.Log.Error().Err(err).Send()
				return
			}

			opts := options.Update()
			opts.SetUpsert(true)
			_, err = mongo.FileOffset.UpdateOne(ctx, bson.D{}, bson.M{
				"$inc": bson.M{
					"index": 1,
				},
			}, opts)
			if err != nil {
				logger.Log.Error().Err(err).Send()
				return
			}
		}
	}
}

func sendLine(line string) (err error) {
	values := strings.Split(line, "\t")

	url := strings.ToLower(values[0])
	registar := strings.ToLower(values[1])
	registrationDate, err := time.Parse("02.01.2006", values[2])
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return
	}

	bytes, err := json.Marshal(protocol.URLMessage{
		URL:              url,
		Registrar:        registar,
		RegistrationDate: registrationDate,
	})
	if err != nil {
		logger.Log.Error().Err(err).Send()
		return
	}

	_, err = stan.Conn.PublishAsync("url", bytes, nil)
	if err != nil {
		logger.Log.Error().Err(err).Send()
	}
	return
}
