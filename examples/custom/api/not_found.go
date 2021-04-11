package api

import (
	"net/http"

	"github.com/wspowell/spiderweb/endpoint"
)

type notFoundResource struct{}

func (self *notFoundResource) Handle(ctx *endpoint.Context) (int, error) {
	return http.StatusNotFound, nil
}
