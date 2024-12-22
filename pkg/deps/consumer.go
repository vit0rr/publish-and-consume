package deps

import (
	"context"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vit0rr/publish-and-consume/config"
	"github.com/vit0rr/publish-and-consume/pkg/log"
	"github.com/vit0rr/publish-and-consume/shared"
	"go.mongodb.org/mongo-driver/mongo"
)

// Message represents the structure of messages received from the queue
type Message struct {
	Key string
}

type Consumer struct {
	deps     *Deps
	RabbitMQ *amqp.Connection
	Mongo    *mongo.Database
}

func NewConsumer(deps *Deps, amqpConn *amqp.Connection, db *mongo.Database, config config.Config, ctx context.Context) (*Consumer, error) {
	return &Consumer{
		deps:     deps,
		RabbitMQ: amqpConn,
		Mongo:    db,
	}, nil
}

func StartConsumers(ctx context.Context, cancel context.CancelFunc, cfg config.Config, mongo *mongo.Client, amqpConn *amqp.Connection, dependencies *Deps) error {
	eventsConsumer, err := NewConsumer(dependencies, amqpConn, mongo.Database("db_events"), cfg, ctx)
	if err != nil {
		log.Error(ctx, "failed to create events consumer", log.ErrAttr(err))
		os.Exit(1)
	}

	go func() {
		if err := eventsConsumer.Start(ctx); err != nil {
			log.Error(ctx, "consumer error", log.ErrAttr(err))
			cancel()
		}
	}()

	return nil
}

func (c *Consumer) Start(ctx context.Context) error {
	ch, err := c.RabbitMQ.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		shared.EventsDataQueue, // queue
		"",                     // consumer
		false,                  // auto-ack
		false,                  // exclusive
		false,                  // no-local
		false,                  // no-wait
		nil,                    // args
	)
	if err != nil {
		return err
	}

	log.Info(ctx, "📋 Started consuming messages from queue", log.AnyAttr("queue", shared.EventsDataQueue))

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}

			key := string(msg.Body)

			// Process the message
			_, err := c.processMessage(ctx, Message{Key: key})
			if err != nil {
				log.Error(ctx, "failed to process message", log.ErrAttr(err))
				msg.Nack(false, true) // Reject message and requeue
				continue
			}

			log.Info(ctx, "Successfully processed message")

			msg.Ack(false) // Acknowledge message
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg Message) (string, error) {
	log.Info(ctx, "Processing message",
		log.AnyAttr("key", msg.Key),
	)

	return "", nil
}
