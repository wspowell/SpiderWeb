package api

import (
	"context"
	"net/http"
)

type notFoundResource struct{}

func (self *notFoundResource) Handle(ctx context.Context) (int, error) {
	return http.StatusNotFound, nil
}
