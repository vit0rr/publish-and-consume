package urlshort

import (
	"net/http"

	"github.com/vit0rr/publish-and-consume/pkg/deps"
	"github.com/vit0rr/publish-and-consume/pkg/log"
	"go.mongodb.org/mongo-driver/mongo"
)

type HTTP struct {
	service *Service
}

func NewHTTP(deps *deps.Deps, db *mongo.Database) *HTTP {
	return &HTTP{
		service: NewService(deps, db),
	}
}

// @summary Short URL
// @description Short URL
// @router /short-url [post]
// @param   body        body        string  true  "URL to short"
// @success 200         {object}    Response                         "Successfully processed transcriptions"
// @failure 400         {object}    Response                         "Invalid request body"
// @failure 500         {object}    Response                         "Internal server error during processing"
func (h *HTTP) ShortUrl(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	url, err := h.service.ShortUrl(r.Context(), r.Body, *h.service.db.Client())
	if err != nil {
		log.Error(r.Context(), "Failed to create short URL", log.ErrAttr(err))
		return nil, err
	}

	return url, nil
}

// @summary Redirect
// @description Redirect
// @router /{id} [get]
// @param   id        path        string  true  "ID"
// @success 200         {object}    Response                         "Successfully processed transcriptions"
// @failure 400         {object}    Response                         "Invalid request body"
// @failure 500         {object}    Response                         "Internal server error during processing"
func (h *HTTP) Redirect(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	url, err := h.service.Redirect(r.Context(), r.URL.Path[1:], *h.service.db.Client())
	if err != nil {
		log.Error(r.Context(), "Failed to redirect", log.ErrAttr(err))
		http.Error(w, "Not found", http.StatusNotFound)
		return nil, err
	}

	http.Redirect(w, r, url, http.StatusMovedPermanently)
	return nil, nil
}
