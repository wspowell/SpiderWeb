package app

import (
	"net/http"

	"spiderweb/endpoint"
	"spiderweb/profiling"
)

type getResource struct {
	Test         string
	ResourceId   int                  `spiderweb:"path=id"`
	ResponseBody *myResponseBodyModel `spiderweb:"response,mime=json,validate"`
}

func (self *getResource) Handle(ctx *endpoint.Context) (int, error) {
	defer profiling.Profile(ctx, "getResource").Finish()
	ctx.Debug("handling getResource")

	ctx.Info("resource id: %v", self.ResourceId)

	self.ResponseBody = &myResponseBodyModel{
		MyString: "test",
		MyInt:    self.ResourceId,
	}

	return http.StatusOK, nil
}
