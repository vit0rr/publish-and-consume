package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// API related config
type API struct {
	RabbitMQ RabbitMQ `hcl:"rabbitmq,block"`
	Mongo    Mongo    `hcl:"mongo,block"`
}

type RabbitMQ struct {
	Uri string `hcl:"uri"`
}

type Mongo struct {
	Dsn string `hcl:"dsn,attr"`
}

func GetDefaultAPIConfig() API {
	godotenv.Load()

	return API{
		Mongo: Mongo{
			Dsn: os.Getenv("DATABASE_URL"),
		},
		RabbitMQ: RabbitMQ{
			Uri: fmt.Sprintf("amqp://%s:%s@%s:%s",
				os.Getenv("RABBIT_MQ_USER"),
				os.Getenv("RABBIT_MQ_PASS"),
				os.Getenv("RABBIT_MQ_HOST"),
				os.Getenv("RABBIT_MQ_PORT"),
			),
		},
	}
}
