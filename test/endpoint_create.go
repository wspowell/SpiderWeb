package test

import (
	"math/rand"
	"time"

	"github.com/wspowell/context"
	"github.com/wspowell/errors"
	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/body"
	"github.com/wspowell/spiderweb/httpstatus"
	"github.com/wspowell/spiderweb/mime"
	"github.com/wspowell/spiderweb/profiling"
	"github.com/wspowell/spiderweb/request"
)

type CreateRequest struct {
	mime.Json

	MyString   string `json:"myString"`
	MyInt      int    `json:"myInt"`
	ShouldFail bool   `json:"shouldFail"`
}

type CreateResponse struct {
	mime.Json

	OutputString string `json:"outputString"`
	OutputInt    int    `json:"outputInt"`
}

type createQueryParams struct {
	ForBench bool
}

func (self *createQueryParams) QueryParameters() []request.Parameter {
	return []request.Parameter{
		request.NewParam("for_bench", &self.ForBench),
	}
}

type Create struct {
	body.Request[CreateRequest]
	body.Response[CreateResponse]
	createQueryParams
}

func (self *Create) Handle(ctx context.Context) (int, error) {
	defer profiling.Profile(ctx, "PostResource").Finish()
	log.Debug(ctx, "handling PostResource")

	if self.RequestBody.ShouldFail {
		return httpstatus.UnprocessableEntity, errors.New("invalid input")
	}

	// If running benchmarks, do not add randomness.
	if !self.ForBench {
		saveResource(ctx)
	}

	self.ResponseBody = CreateResponse{
		OutputString: self.RequestBody.MyString,
		OutputInt:    self.RequestBody.MyInt,
	}

	return httpstatus.Created, nil
}

// Fake spending time to save data.
func saveResource(ctx context.Context) {
	defer profiling.Profile(ctx, "saveResource").Finish()

	source := rand.NewSource(time.Now().UnixNano())
	// nolint:gosec // reason: no need for secure random here
	random := rand.New(source)

	time.Sleep(time.Duration(random.Intn(500)) * time.Millisecond)
}
