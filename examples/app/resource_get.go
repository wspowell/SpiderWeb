package app

import (
	"net/http"

	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/profiling"
)

type GetResource struct {
	Test         string
	ResourceId   int                  `spiderweb:"path=id"`
	ResponseBody *MyResponseBodyModel `spiderweb:"response,mime=application/json,validate"`
}

func (self *GetResource) Handle(ctx *endpoint.Context) (int, error) {
	defer profiling.Profile(ctx, "GetResource").Finish()
	ctx.Debug("handling GetResource")

	ctx.Info("resource id: %v", self.ResourceId)

	self.ResponseBody = &MyResponseBodyModel{
		MyString: "test",
		MyInt:    self.ResourceId,
	}

	return http.StatusOK, nil
}
