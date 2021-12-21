package restful_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/httpmethod"
	"github.com/wspowell/spiderweb/httpstatus"
	"github.com/wspowell/spiderweb/server/restful"
	"github.com/wspowell/spiderweb/server/route"
	"github.com/wspowell/spiderweb/test"
)

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
	config := &endpoint.Config{
		LogConfig: &test.NoopLogConfig{
			Config: log.NewConfig().WithLevel(log.LevelFatal),
		},
		Resources: map[string]any{
			"datastore": &test.Database{},
		},
		Timeout: 30 * time.Second,
	}

	sample.HandleNotFound(config, &test.NoRoute{})
	sample.Handle(config, route.Post("/sample", &test.Create{}))
	sample.Handle(config, route.Get("/sample/{id}", &test.Get{}))
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
		httpStatus, _ := sample.Execute(&requestCtx)
		if httpStatus != httpstatus.Created {
			panic(fmt.Sprintf("status not 201: %v", httpStatus))
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
			httpStatus, _ := sample.Execute(&requestCtx)
			if httpStatus != httpstatus.Created {
				panic(fmt.Sprintf("status not 201: %v", httpStatus))
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
		httpStatus, _ := sample.Execute(&requestCtx)
		if httpStatus != httpstatus.OK {
			panic("status not 200")
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
