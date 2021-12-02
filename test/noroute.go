package test

import (
	"github.com/wspowell/context"

	"github.com/wspowell/spiderweb/httpstatus"
)

type NoRoute struct{}

func (self *NoRoute) Handle(ctx context.Context) (int, error) {
	return httpstatus.NotFound, nil
}
