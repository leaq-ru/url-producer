package mongo

import (
	"context"
	"github.com/nnqq/scr-url-producer/config"
	"github.com/nnqq/scr-url-producer/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"time"
)

var FileEntity *mongo.Collection

const (
	db         = "url_producer"
	fileEntity = "file_entity"
)

func init() {
	const timeout = 10
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().
		SetWriteConcern(writeconcern.New(
			writeconcern.W(1),
			writeconcern.J(true),
		)).
		SetReadConcern(readconcern.Available()).
		SetReadPreference(readpref.SecondaryPreferred()).
		ApplyURI(config.Env.MongoDB.URL))
	logger.Must(err)

	err = client.Ping(ctx, nil)
	logger.Must(err)

	urlProducer := client.Database(db)
	FileEntity = urlProducer.Collection(fileEntity)
}
