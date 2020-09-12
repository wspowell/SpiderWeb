package app

import (
	"math/rand"
	"net/http"
	"time"

	"spiderweb/endpoint"
	"spiderweb/errors"
	"spiderweb/local"
	"spiderweb/profiling"
)

type postResource struct {
	Test         string
	RequestBody  *myRequestBodyModel  `spiderweb:"request,mime=json,validate"`
	ResponseBody *myResponseBodyModel `spiderweb:"response,mime=json,validate"`
}

func (self *postResource) Handle(ctx *endpoint.Context) (int, error) {
	defer profiling.Profile(ctx, "postResource").Finish()
	ctx.Debug("handling postResource")

	if self.RequestBody.ShouldFail {
		return http.StatusUnprocessableEntity, errors.New("APP1234", "invalid input")
	}

	saveResource(ctx)

	self.ResponseBody = &myResponseBodyModel{
		MyString: self.RequestBody.MyString,
		MyInt:    self.RequestBody.MyInt,
	}

	return http.StatusCreated, nil
}

// Fake spending time to save data.
func saveResource(ctx local.Context) {
	defer profiling.Profile(ctx, "saveResource").Finish()

	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	time.Sleep(time.Duration(random.Intn(500)) * time.Millisecond)
}
