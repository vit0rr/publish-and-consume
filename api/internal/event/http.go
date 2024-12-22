package event

import (
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vit0rr/publish-and-consume/pkg/deps"
	"github.com/vit0rr/publish-and-consume/pkg/log"
	"go.mongodb.org/mongo-driver/mongo"
)

type HTTP struct {
	service *Service
}

func NewHTTP(deps *deps.Deps, db *mongo.Database, amqpConn *amqp.Connection) *HTTP {
	return &HTTP{
		service: NewService(deps, db, amqpConn),
	}
}

// @summary Publish to queue
// @description Publish to queue
// @router / [post]
// @param   body        body        PublishToQueueBody  true  "Publish to queue body"
// @success 200         {object}    Response                         "Successfully published to queue"
// @failure 400         {object}    Response                         "Invalid request body"
// @failure 500         {object}    Response                         "Internal server error during processing"
func (h *HTTP) PublishToQueue(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	msg, err := h.service.PublishToQueue(r.Context(), r.Body, *h.service.db.Client())
	if err != nil {
		log.Error(r.Context(), "Failed to publish to queue", log.ErrAttr(err))
		return nil, err
	}

	return msg, nil
}

type PublishToQueueBody struct {
	Username string `json:"username"`
}
