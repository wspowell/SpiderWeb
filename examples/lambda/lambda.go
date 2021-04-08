package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/wspowell/errors"
	"github.com/wspowell/local"
	"github.com/wspowell/logging"
	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/lambda"
	"github.com/wspowell/spiderweb/profiling"
)

type MyRequestBodyModel struct {
	MyString   string `json:"my_string"`
	MyInt      int    `json:"my_int"`
	ShouldFail bool   `json:"fail"`
}

type MyResponseBodyModel struct {
	MyString string `json:"output_string"`
	MyInt    int    `json:"output_int"`
}

type PostResource struct {
	Test         string
	RequestBody  *MyRequestBodyModel  `spiderweb:"request,mime=application/json,validate"`
	ResponseBody *MyResponseBodyModel `spiderweb:"response,mime=application/json,validate"`
}

func (self *PostResource) Handle(ctx *endpoint.Context) (int, error) {
	defer profiling.Profile(ctx, "PostResource").Finish()
	ctx.Debug("handling PostResource")

	if self.RequestBody.ShouldFail {
		return http.StatusUnprocessableEntity, errors.New("APP1234", "invalid input")
	}

	saveResource(ctx)

	self.ResponseBody = &MyResponseBodyModel{
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

func main() {
	endpointConfig := &endpoint.Config{
		LogConfig: logging.NewConfig(logging.LevelDebug, map[string]interface{}{}),
		Timeout:   30 * time.Second,
	}

	handler := lambda.NewLambda(endpointConfig, &PostResource{})
	handler.Start()
}
