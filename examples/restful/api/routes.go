package api

import (
	"github.com/wspowell/context"

	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/examples/restful/api/items"
	"github.com/wspowell/spiderweb/examples/restful/resources"
	"github.com/wspowell/spiderweb/handler"
	"github.com/wspowell/spiderweb/httpstatus"
	"github.com/wspowell/spiderweb/server/restful"
)

func Routes(custom *restful.Server, datastore resources.Datastore) {
	custom.HandleNotFound(handler.NewHandle(noRoute{}).
		WithLogConfig(log.NewConfig().WithLevel(log.LevelDebug)))

	items.Routes(custom, datastore)
}

type noRoute struct{}

func (self *noRoute) Handle(ctx context.Context) (int, error) {
	return httpstatus.NotFound, nil
}
