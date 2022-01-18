package items

import (
	"time"

	"github.com/wspowell/spiderweb/examples/restful/middleware"
	"github.com/wspowell/spiderweb/examples/restful/resources"
	"github.com/wspowell/spiderweb/handler"
	"github.com/wspowell/spiderweb/httpmethod"
	"github.com/wspowell/spiderweb/server/restful"
)

func Routes(custom *restful.Server, database resources.Datastore) {
	custom.Handle(httpmethod.Post, "/items", handler.NewHandle(postResource{}).
		WithErrorResponse(middleware.AllErrorsTeapot))
	custom.Handle(httpmethod.Get, "/items/{id}", handler.NewHandle(getResource{
		Db: database,
	}).WithETag(30*time.Second))
}
