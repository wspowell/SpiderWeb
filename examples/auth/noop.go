package auth

import (
	"net/http"

	"github.com/wspowell/spiderweb/endpoint"
)

type Noop struct{}

func (self Noop) Auth(ctx *endpoint.Context, VisitAllHeaders func(func(key, value []byte))) (int, error) {
	return http.StatusOK, nil
}
