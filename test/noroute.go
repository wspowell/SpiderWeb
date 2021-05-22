package test

import (
	"github.com/wspowell/context"
	"github.com/wspowell/spiderweb/http"
)

type NoRoute struct{}

func (self *NoRoute) Handle(ctx context.Context) (int, error) {
	return http.StatusNotFound, nil
}
