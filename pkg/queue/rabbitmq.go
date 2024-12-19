package queue

import (
	"context"
	"fmt"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vit0rr/publish-and-consume/config"
	"github.com/vit0rr/publish-and-consume/pkg/consumer/events"
	"github.com/vit0rr/publish-and-consume/pkg/deps"
	"github.com/vit0rr/publish-and-consume/pkg/log"
	"github.com/vit0rr/publish-and-consume/shared"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewRabbitMQClient(
	ctx context.Context,
	cfg config.Config,
) (*amqp.Connection, error) {
	amqpConn, err := amqp.Dial(cfg.API.RabbitMQ.Uri)
	if err != nil {
		return nil, err
	}

	return amqpConn, nil
}

func DeclareQueuesAndExanghes(cnn *amqp.Connection) error {
	ch, err := cnn.Channel()
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}
	defer ch.Close()

	// Declare exchanges
	if err := ch.ExchangeDeclare(
		shared.EventsTopic, // name
		"fanout",           // type
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	); err != nil {
		return fmt.Errorf("failed to declare exchange %s: %w", shared.EventsTopic, err)
	}

	if err := ch.ExchangeDeclare(
		shared.EventsDLT, // name
		"fanout",         // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	); err != nil {
		return fmt.Errorf("failed to declare exchange %s: %w", shared.EventsDLT, err)
	}

	// Declare queues
	if _, err := ch.QueueDeclare(
		shared.EventsDataQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    shared.EventsDLT,
			"x-dead-letter-routing-key": shared.EventsDataDQL,
		},
	); err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", shared.EventsDataQueue, err)
	}

	if _, err := ch.QueueDeclare(
		shared.EventsDataDQL,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-dead-letter-exchange": shared.EventsDLT,
		},
	); err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", shared.EventsDataDQL, err)
	}

	// Bind queues to exchanges
	if err := ch.QueueBind(
		shared.EventsDataQueue,
		"",                 // routing key
		shared.EventsTopic, // exchange
		false,              // no-wait
		nil,                // arguments
	); err != nil {
		return fmt.Errorf("failed to bind queue %s to exchange %s: %w",
			shared.EventsDataQueue, shared.EventsTopic, err)
	}

	if err := ch.QueueBind(
		shared.EventsDataDQL,
		"",               // routing key
		shared.EventsDLT, // exchange
		false,            // no-wait
		nil,              // arguments
	); err != nil {
		return fmt.Errorf("failed to bind queue %s to exchange %s: %w",
			shared.EventsDataDQL, shared.EventsDLT, err)
	}

	return nil
}

func StartConsumers(ctx context.Context, cancel context.CancelFunc, cfg config.Config, mongo *mongo.Client, amqpConn *amqp.Connection, dependencies *deps.Deps) error {
	eventsConsumer, err := events.NewConsumer(dependencies, amqpConn, mongo.Database("db_events"), cfg, ctx)
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
