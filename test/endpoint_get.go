package test

import (
	"github.com/wspowell/context"
	"github.com/wspowell/log"
	"github.com/wspowell/spiderweb/http"
	"github.com/wspowell/spiderweb/profiling"
)

type fooResponseModel struct {
	MyString string `json:"output_string"`
	MyInt    int    `json:"output_int"`
}

type Get struct {
	Test         string
	Db           Datastore         `spiderweb:"resource=datastore"`
	ResourceId   int               `spiderweb:"path=id"`
	ResponseBody *fooResponseModel `spiderweb:"response,mime=application/json,validate"`
}

func (self *Get) Handle(ctx context.Context) (int, error) {
	defer profiling.Profile(ctx, "GetResource").Finish()
	log.Debug(ctx, "handling GetResource")

	log.Info(ctx, "resource id: %v", self.ResourceId)

	self.ResponseBody = &fooResponseModel{
		MyString: self.Db.RetrieveValue(),
		MyInt:    self.ResourceId,
	}

	return http.StatusOK, nil
}
