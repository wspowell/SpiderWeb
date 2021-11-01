package middleware

import (
	"net/http"

	"github.com/wspowell/context"
)

type AuthNoop struct{}

func (self AuthNoop) Authorization(ctx context.Context, PeekHeader func(key string) []byte) (int, error) {
	return http.StatusOK, nil
}
