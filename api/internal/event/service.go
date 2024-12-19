package urlshort

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vit0rr/publish-and-consume/pkg/deps"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service struct {
	deps     *deps.Deps
	db       *mongo.Database
	amqpConn *amqp.Connection
}

type Response struct {
	Message string `json:"message"`
}

func NewService(deps *deps.Deps, db *mongo.Database, amqpConn *amqp.Connection) *Service {
	return &Service{
		deps:     deps,
		db:       db,
		amqpConn: amqpConn,
	}
}

// Publish to queue
func (s *Service) PublishToQueue(c context.Context, b io.ReadCloser, dbclient mongo.Client) (*Response, error) {
	var body PublishToQueueBody
	if err := json.NewDecoder(b).Decode(&body); err != nil {
		return nil, err
	}

	// TODO: Publish to queue using RabbitMQ

	return &Response{
		Message: fmt.Sprintf("Published to queue: %s", body.Username),
	}, nil
}
