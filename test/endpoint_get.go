package test

import (
	"github.com/wspowell/context"
	"github.com/wspowell/log"
	"github.com/wspowell/spiderweb/body"
	"github.com/wspowell/spiderweb/httpstatus"
	"github.com/wspowell/spiderweb/mime"
	"github.com/wspowell/spiderweb/profiling"
	"github.com/wspowell/spiderweb/request"
)

type fooResponseModel struct {
	mime.Json

	OutputString string `json:"outputString"`
	OutputInt    int    `json:"outputInt"`
}

type getPathParams struct {
	ResourceId int
}

func (self *getPathParams) PathParameters() []request.Parameter {
	return []request.Parameter{
		request.NewParam("id", &self.ResourceId),
	}
}

type Get struct {
	Db Datastore
	body.Response[fooResponseModel]
	getPathParams
}

func (self *Get) Handle(ctx context.Context) (int, error) {
	defer profiling.Profile(ctx, "GetResource").Finish()
	log.Debug(ctx, "handling GetResource")

	log.Info(ctx, "resource id: %v", self.ResourceId)

	value := self.Db.RetrieveValue()
	log.Debug(ctx, "retrieved value: %s", value)

	self.ResponseBody = fooResponseModel{
		OutputString: value,
		OutputInt:    self.ResourceId,
	}

	return httpstatus.OK, nil
}
