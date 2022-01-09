package main

import (
	"net/http"

	"github.com/wspowell/context"

	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/profiling"
)

type fooResponseModel struct {
	MyString string `json:"outputString"`
	MyInt    int    `json:"outputInt"`
}

type get struct {
	Test         string
	ResourceId   int               `spiderweb:"path=id"`
	ResponseBody *fooResponseModel `spiderweb:"response,mime=application/json,validate"`
}

func (self *get) Handle(ctx context.Context) (int, error) {
	defer profiling.Profile(ctx, "GetResource").Finish()
	log.Debug(ctx, "handling GetResource")

	log.Info(ctx, "resource id: %v", self.ResourceId)

	self.ResponseBody = &fooResponseModel{
		MyString: "test",
		MyInt:    self.ResourceId,
	}

	return http.StatusOK, nil
}
