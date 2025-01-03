package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	amqp "github.com/rabbitmq/amqp091-go"
	httpSwagger "github.com/swaggo/http-swagger"
	event "github.com/vit0rr/publish-and-consume/api/internal/event"
	_ "github.com/vit0rr/publish-and-consume/docs" // docs is generated by Swag CLI, you have to import it.
	"github.com/vit0rr/publish-and-consume/pkg/deps"
	"github.com/vit0rr/publish-and-consume/pkg/telemetry"
	"go.mongodb.org/mongo-driver/mongo"
)

type Router struct {
	Deps  *deps.Deps
	event *event.HTTP
}

func (router *Router) BuildRoutes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.StripSlashes)

	// Custom middlewares
	r.Use(telemetry.TelemetryMiddleware)
	r.Use(SetResponseTypeToJSON)
	r.Use(CorsMiddleware)
	r.Use(AuthMiddleware)

	r.Route("/", func(r chi.Router) {
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("http://localhost:8080/swagger/doc.json"), //The url pointing to API definition
		))
		r.Post("/", telemetry.HandleFuncLogger(router.event.PublishToQueue))

	})

	return r
}

func New(deps *deps.Deps, db mongo.Database, amqpConn *amqp.Connection) *Router {
	return &Router{
		Deps:  deps,
		event: event.NewHTTP(deps, &db, amqpConn),
	}
}
