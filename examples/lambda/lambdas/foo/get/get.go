package main

import (
	"net/http"

	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/profiling"
)

type fooResponseModel struct {
	MyString string `json:"output_string"`
	MyInt    int    `json:"output_int"`
}

type get struct {
	Test         string
	ResourceId   int               `spiderweb:"path=id"`
	ResponseBody *fooResponseModel `spiderweb:"response,mime=application/json,validate"`
}

func (self *get) Handle(ctx *endpoint.Context) (int, error) {
	defer profiling.Profile(ctx, "GetResource").Finish()
	ctx.Debug("handling GetResource")

	ctx.Info("resource id: %v", self.ResourceId)

	self.ResponseBody = &fooResponseModel{
		MyString: "test",
		MyInt:    self.ResourceId,
	}

	return http.StatusOK, nil
}
