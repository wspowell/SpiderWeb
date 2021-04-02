package app

import (
	"net/http"

	"github.com/wspowell/spiderweb/endpoint"
)

type NotFoundResource struct{}

func (self *NotFoundResource) Handle(ctx *endpoint.Context) (int, error) {
	return http.StatusNotFound, nil
}
