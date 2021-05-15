package test

import (
	"github.com/wspowell/context"
	"github.com/wspowell/spiderweb/http"
)

type noRoute struct{}

func (self *noRoute) Handle(ctx context.Context) (int, error) {
	return http.StatusNotFound, nil
}
