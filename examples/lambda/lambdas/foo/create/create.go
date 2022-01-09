package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/wspowell/context"

	"github.com/wspowell/errors"
	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/profiling"
)

type createRequest struct {
	MyString   string `json:"myString"`
	MyInt      int    `json:"myInt"`
	ShouldFail bool   `json:"shouldFail"`
}

type createResponse struct {
	MyString string `json:"outputString"`
	MyInt    int    `json:"outputInt"`
}

type create struct {
	Test         string
	RequestBody  *createRequest  `spiderweb:"request,mime=application/json,validate"`
	ResponseBody *createResponse `spiderweb:"response,mime=application/json,validate"`
}

func (self *create) Handle(ctx context.Context) (int, error) {
	defer profiling.Profile(ctx, "PostResource").Finish()
	log.Debug(ctx, "handling PostResource")

	if self.RequestBody.ShouldFail {
		return http.StatusUnprocessableEntity, errors.New("invalid input")
	}

	saveResource(ctx)

	self.ResponseBody = &createResponse{
		MyString: self.RequestBody.MyString,
		MyInt:    self.RequestBody.MyInt,
	}

	return http.StatusCreated, nil
}

// Fake spending time to save data.
func saveResource(ctx context.Context) {
	defer profiling.Profile(ctx, "saveResource").Finish()

	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	time.Sleep(time.Duration(random.Intn(500)) * time.Millisecond)
}
