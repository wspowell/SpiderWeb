package test

import (
	"math/rand"
	"time"

	"github.com/wspowell/context"
	"github.com/wspowell/errors"
	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/httpstatus"
	"github.com/wspowell/spiderweb/profiling"
)

type CreateRequest struct {
	MyString   string `json:"myString"`
	MyInt      int    `json:"myInt"`
	ShouldFail bool   `json:"shouldFail"`
}

type CreateResponse struct {
	OutputString string `json:"outputString"`
	OutputInt    int    `json:"outputInt"`
}

type Create struct {
	Test         string
	ForBench     bool            `spiderweb:"query=for_bench"`
	RequestBody  *CreateRequest  `spiderweb:"request,mime=application/json,validate"`
	ResponseBody *CreateResponse `spiderweb:"response,mime=application/json,validate"`
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

	self.ResponseBody = &CreateResponse{
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
