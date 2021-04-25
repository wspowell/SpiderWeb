package middleware

import (
	"net/http"

	"github.com/wspowell/context"
)

type AuthNoop struct{}

func (self AuthNoop) Auth(ctx context.Context, VisitAllHeaders func(func(key, value []byte))) (int, error) {
	return http.StatusOK, nil
}
