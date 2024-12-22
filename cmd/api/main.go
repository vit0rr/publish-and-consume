package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
	"github.com/vit0rr/publish-and-consume/api/server"
	"github.com/vit0rr/publish-and-consume/cmd/consumer"
	"github.com/vit0rr/publish-and-consume/config"
	_ "github.com/vit0rr/publish-and-consume/docs"
	"github.com/vit0rr/publish-and-consume/pkg/deps"
	"github.com/vit0rr/publish-and-consume/pkg/log"
)

// @title Publish and Consume
// @version 1.0
// @description Publish and Consume with RabbitMQ and MongoDB
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
func main() {
	godotenv.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to the config file (see config/config_local.hcl for an example)")
	flag.Parse()

	var cfg config.Config
	if configPath == "" {
		cfg = config.DefaultConfig()
	} else if configPath != "" {
		parseConfig, err := config.GetConfig(configPath)
		if err != nil {
			log.Error(ctx, "Error parsing config file", "error", err)
			os.Exit(1)
		}

		cfg = parseConfig
	}

	logLevel, err := log.ParseLogLevel(cfg.Server.LogLevel)
	if err != nil {
		log.Error(ctx, "Error parsing log level", "error", err)
		logLevel = slog.LevelInfo
	}
	log.New(ctx, logLevel)

	// implement rabbitmq connection
	amqpConn, err := deps.NewRabbitMQClient(ctx, cfg)
	if err != nil {
		log.Error(ctx, "❌ Failed to connect to RabbitMQ", log.ErrAttr(err))
		os.Exit(1)
	}

	log.Info(ctx, "✅ Connected to RabbitMQ")

	// create mongo client
	mongoClient, err := deps.NewMongoClient(ctx, cfg)
	if err != nil {
		log.Error(ctx, "❌ Unable to parse database connection", log.ErrAttr(err))
		os.Exit(1)
	}

	log.Info(ctx, "✅ Connected to MongoDB")

	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Error(ctx, "❌ Failed to disconnect from MongoDB", log.ErrAttr(err))
			os.Exit(1)
		}
	}()

	// create queues
	if err := deps.DeclareQueuesAndExanghes(amqpConn); err != nil {
		log.Error(ctx, "❌ Failed to declare queues", log.ErrAttr(err))
		os.Exit(1)
	}

	log.Info(ctx, "✅ Created queues")

	defer func() {
		if err := amqpConn.Close(); err != nil {
			log.Error(ctx, "❌ Failed to close RabbitMQ connection", log.ErrAttr(err))
			os.Exit(1)
		}
	}()

	dependencies := deps.New(cfg, mongoClient, amqpConn)

	httpServer := server.New(ctx, dependencies, mongoClient.Database("db_events"), amqpConn)

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := httpServer.Shutdown(ctx); err != nil {
			log.Error(ctx, "unexpected error during server shutdown", log.ErrAttr(err))
		}
		close(idleConnsClosed)
	}()

	if err := consumer.StartConsumers(ctx, cancel, cfg, mongoClient, amqpConn, dependencies); err != nil {
		log.Error(ctx, "failed to start consumers", log.ErrAttr(err))
		os.Exit(1)
	}

	log.Info(ctx, "🌎 Starting API at", log.AnyAttr("bind_addr", cfg.Server.BindAddr))
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error(ctx, "❌ Error starting server", log.ErrAttr(err))
	}

	<-idleConnsClosed

}
