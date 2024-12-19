package deps

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vit0rr/publish-and-consume/config"
	"go.mongodb.org/mongo-driver/mongo"
)

type Deps struct {
	Config   config.Config
	DBClient *mongo.Client
	AmqpConn *amqp.Connection
}

func New(config config.Config, mgClient *mongo.Client, amqpConn *amqp.Connection) *Deps {
	return &Deps{
		Config:   config,
		DBClient: mgClient,
		AmqpConn: amqpConn,
	}
}
