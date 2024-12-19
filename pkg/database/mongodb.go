package database

import (
	"context"

	"github.com/vit0rr/publish-and-consume/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoClient(ctx context.Context, cfg config.Config) (*mongo.Client, error) {
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.API.Mongo.Dsn))
	if err != nil {
		return nil, err
	}

	return mongoClient, nil
}
