package api

import (
	"net/http"

	"github.com/wspowell/context"
)

type notFoundResource struct{}

func (self *notFoundResource) Handle(ctx context.Context) (int, error) {
	return http.StatusNotFound, nil
}
