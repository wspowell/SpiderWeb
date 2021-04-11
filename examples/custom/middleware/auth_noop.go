package middleware

import (
	"net/http"

	"github.com/wspowell/spiderweb/endpoint"
)

type AuthNoop struct{}

func (self AuthNoop) Auth(ctx *endpoint.Context, VisitAllHeaders func(func(key, value []byte))) (int, error) {
	return http.StatusOK, nil
}
