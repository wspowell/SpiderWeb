package restful_test

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/wspowell/context"
	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/body"
	"github.com/wspowell/spiderweb/handler"
	"github.com/wspowell/spiderweb/httpmethod"
	"github.com/wspowell/spiderweb/httpstatus"
	"github.com/wspowell/spiderweb/mime"
	"github.com/wspowell/spiderweb/profiling"
	"github.com/wspowell/spiderweb/request"
	"github.com/wspowell/spiderweb/server/restful"
	"github.com/wspowell/spiderweb/test"
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

type testCreate struct {
	body.Request[CreateRequest]
	body.Response[CreateResponse]
	createQueryParams
}

func (self *testCreate) Handle(ctx context.Context) (int, error) {
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

type testGet struct {
	Db test.Datastore
	body.Response[fooResponseModel]
	getPathParams
}

func (self *testGet) Handle(ctx context.Context) (int, error) {
	defer profiling.Profile(ctx, "GetResource").Finish()
	log.Debug(ctx, "handling GetResource")

	log.Info(ctx, "resource id: %v", self.ResourceId)

	self.ResponseBody = fooResponseModel{
		OutputString: self.Db.RetrieveValue(),
		OutputInt:    self.ResourceId,
	}

	return httpstatus.OK, nil
}

func routes() *restful.Server {
	serverConfig := &restful.ServerConfig{
		LogConfig: &test.NoopLogConfig{
			Config: log.NewConfig().WithLevel(log.LevelFatal),
		},
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		EnablePprof:  false,
	}

	sample := restful.NewServer(serverConfig)

	sampleRoutes(sample)

	return sample
}

func sampleRoutes(sample *restful.Server) {
	sample.HandleNotFound(handler.NewHandle(test.NoRoute{}))
	sample.Handle(httpmethod.Post, "/sample", handler.NewHandle(testCreate{}))
	sample.Handle(httpmethod.Get, "/sample/{id}", handler.NewHandle(testGet{
		Db: &test.Database{},
	}))
}

func Benchmark_SpiderWeb_POST_latency(b *testing.B) {
	sample := routes()

	var req fasthttp.Request

	req.Header.SetMethod(httpmethod.Post)
	req.Header.SetRequestURI("/sample?for_bench=true")
	req.Header.Set(fasthttp.HeaderHost, "localhost")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBody([]byte(`{"outputString":"hello","outputInt":5}`))

	requestCtx := fasthttp.RequestCtx{}
	requestCtx.Init(&req, nil, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		httpStatus, responseBytes := sample.Execute(&requestCtx)
		if httpStatus != httpstatus.Created {
			panic(fmt.Sprintf("status not 201: %v, %s", httpStatus, string(responseBytes)))
		}
	}
}

func Benchmark_SpiderWeb_POST_throughput(b *testing.B) {
	sample := routes()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var req fasthttp.Request

		req.Header.SetMethod(httpmethod.Post)
		req.Header.SetRequestURI("/sample?for_bench=true")
		req.Header.Set(fasthttp.HeaderHost, "localhost")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.SetBody([]byte(`{"outputString":"hello","outputInt":5}`))

		requestCtx := fasthttp.RequestCtx{}
		requestCtx.Init(&req, nil, nil)

		for pb.Next() {
			httpStatus, responseBytes := sample.Execute(&requestCtx)
			if httpStatus != httpstatus.Created {
				panic(fmt.Sprintf("status not 201: %v, %s", httpStatus, string(responseBytes)))
			}
		}
	})
}

func Benchmark_SpiderWeb_GET_latency(b *testing.B) {
	sample := routes()

	var req fasthttp.Request

	req.Header.SetMethod(httpmethod.Get)
	req.Header.SetRequestURI("/sample/34")
	req.Header.Set(fasthttp.HeaderHost, "localhost")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	requestCtx := fasthttp.RequestCtx{}
	requestCtx.Init(&req, nil, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		httpStatus, responseBytes := sample.Execute(&requestCtx)
		if httpStatus != httpstatus.OK {
			panic(fmt.Sprintf("status not 200: %v %s", httpStatus, string(responseBytes)))
		}
	}
}

func Benchmark_SpiderWeb_GET_throughput(b *testing.B) {
	sample := routes()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var req fasthttp.Request

		req.Header.SetMethod(httpmethod.Get)
		req.Header.SetRequestURI("/sample/34")
		req.Header.Set(fasthttp.HeaderHost, "localhost")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		requestCtx := fasthttp.RequestCtx{}
		requestCtx.Init(&req, nil, nil)

		for pb.Next() {
			httpStatus, _ := sample.Execute(&requestCtx)
			if httpStatus != httpstatus.OK {
				panic("status not 200")
			}
		}
	})
}
