package resource

import (
	"net/http"

	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/profiling"
)

type getResource struct {
	Test         string
	Db           Datastore            `spiderweb:"resource=datastore"`
	ResourceId   int                  `spiderweb:"path=id"`
	ResponseBody *MyResponseBodyModel `spiderweb:"response,mime=application/json,validate"`
}

func (self *getResource) Handle(ctx *endpoint.Context) (int, error) {
	defer profiling.Profile(ctx, "GetResource").Finish()
	ctx.Debug("handling GetResource")

	ctx.Info("resource id: %v", self.ResourceId)

	self.Db.RetrieveValue()

	self.ResponseBody = &MyResponseBodyModel{
		MyString: "test",
		MyInt:    self.ResourceId,
	}

	return http.StatusOK, nil
}

type Datastore interface {
	RetrieveValue() string
}

var _ Datastore = (*Database)(nil)

type Database struct{}

func (self *Database) RetrieveValue() string {
	panic("external call")
}
