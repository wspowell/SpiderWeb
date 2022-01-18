package items

import (
	"net/http"

	"github.com/wspowell/context"
	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/body"
	"github.com/wspowell/spiderweb/examples/restful/resources"
	"github.com/wspowell/spiderweb/profiling"
	"github.com/wspowell/spiderweb/request"
)

type getResource struct {
	body.Response[MyResponseBodyModel]
	ResourceId int

	Db resources.Datastore
}

func (self *getResource) PathParameters() []request.Parameter {
	return []request.Parameter{
		request.NewParam("id", &self.ResourceId),
	}
}

func (self *getResource) Handle(ctx context.Context) (int, error) {
	defer profiling.Profile(ctx, "GetResource").Finish()
	log.Debug(ctx, "handling GetResource")

	log.Info(ctx, "resource id: %v", self.ResourceId)

	self.Db.RetrieveValue()

	self.ResponseBody = MyResponseBodyModel{
		MyString: "test",
		MyInt:    self.ResourceId,
	}

	return http.StatusOK, nil
}

var _ resources.Datastore = (*Database)(nil)

type Database struct{}

func (self *Database) RetrieveValue() string {
	panic("external call")
}
