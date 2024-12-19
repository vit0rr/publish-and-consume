package server

import (
	"context"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vit0rr/publish-and-consume/api/router"
	"github.com/vit0rr/publish-and-consume/pkg/deps"
	"go.mongodb.org/mongo-driver/mongo"
)

func New(ctx context.Context, deps *deps.Deps, db *mongo.Database, amqpConn *amqp.Connection) *http.Server {
	router := router.New(deps, *db, amqpConn)
	timeout := time.Duration(deps.Config.Server.CtxTimeout) * time.Second

	return &http.Server{
		Addr:         deps.Config.Server.BindAddr,
		Handler:      http.TimeoutHandler(router.BuildRoutes(), timeout, ""),
		ReadTimeout:  timeout,
		WriteTimeout: timeout + 2,
	}
}
