package servertest

import (
	"net/http"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/wspowell/context"
	"github.com/wspowell/log"
	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/server"
)

type benchRequest struct {
	MyString   string `json:"my_string"`
	MyInt      int    `json:"my_int"`
	ShouldFail bool   `json:"fail"`
}

type benchResponse struct {
	MyString string `json:"output_string"`
	MyInt    int    `json:"output_int"`
}

type bench struct {
	Test string

	RequestBody  *createRequest  `spiderweb:"request,mime=application/json,validate"`
	ResponseBody *createResponse `spiderweb:"response,mime=application/json,validate"`
}

func (self *bench) Handle(ctx context.Context) (int, error) {
	self.ResponseBody = &createResponse{
		MyString: self.RequestBody.MyString,
		MyInt:    self.RequestBody.MyInt,
	}

	return http.StatusCreated, nil
}

func Benchmark_SpiderWeb_POST_latency(b *testing.B) {
	serverConfig := &server.Config{
		LogConfig: &noopLogConfig{
			Config: log.NewConfig(log.LevelFatal),
		},
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		EnablePprof:  false,
	}

	sample := server.New(serverConfig)

	config := &endpoint.Config{
		LogConfig: &noopLogConfig{
			Config: log.NewConfig(log.LevelFatal),
		},
		Resources: map[string]interface{}{
			"datastore": &database{},
		},
		Timeout: 30 * time.Second,
	}

	sample.Handle(config, http.MethodPost, "/bench", &bench{})

	var req fasthttp.Request

	req.Header.SetMethod(http.MethodPost)
	req.Header.SetRequestURI("/bench")
	req.Header.Set(fasthttp.HeaderHost, "localhost")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBody([]byte(`{"output_string":"hello","output_int":5}`))

	requestCtx := fasthttp.RequestCtx{}
	requestCtx.Init(&req, nil, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		httpStatus, _ := sample.Execute(&requestCtx)
		if httpStatus != http.StatusCreated {
			panic("status not 201")
		}
	}
}

func Benchmark_SpiderWeb_POST_throughput(b *testing.B) {

	serverConfig := &server.Config{
		LogConfig: &noopLogConfig{
			Config: log.NewConfig(log.LevelFatal),
		},
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		EnablePprof:  false,
	}

	sample := server.New(serverConfig)

	config := &endpoint.Config{
		LogConfig: &noopLogConfig{
			Config: log.NewConfig(log.LevelFatal),
		},
		Resources: map[string]interface{}{
			"datastore": &database{},
		},
		Timeout: 30 * time.Second,
	}

	sample.Handle(config, http.MethodPost, "/bench", &bench{})

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var req fasthttp.Request

		req.Header.SetMethod(http.MethodPost)
		req.Header.SetRequestURI("/bench")
		req.Header.Set(fasthttp.HeaderHost, "localhost")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.SetBody([]byte(`{"output_string":"hello","output_int":5}`))

		requestCtx := fasthttp.RequestCtx{}
		requestCtx.Init(&req, nil, nil)

		for pb.Next() {
			httpStatus, _ := sample.Execute(&requestCtx)
			if httpStatus != http.StatusCreated {
				panic("status not 201")
			}
		}
	})
}

func Benchmark_SpiderWeb_GET_latency(b *testing.B) {

	sample := routes()

	var req fasthttp.Request

	req.Header.SetMethod(http.MethodGet)
	req.Header.SetRequestURI("/sample/34")
	req.Header.Set(fasthttp.HeaderHost, "localhost")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	requestCtx := fasthttp.RequestCtx{}
	requestCtx.Init(&req, nil, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		httpStatus, _ := sample.Execute(&requestCtx)
		if httpStatus != http.StatusOK {
			panic("status not 200")
		}
	}
}

func Benchmark_SpiderWeb_GET_throughput(b *testing.B) {

	sample := routes()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var req fasthttp.Request

		req.Header.SetMethod(http.MethodGet)
		req.Header.SetRequestURI("/sample/34")
		req.Header.Set(fasthttp.HeaderHost, "localhost")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		requestCtx := fasthttp.RequestCtx{}
		requestCtx.Init(&req, nil, nil)

		for pb.Next() {
			httpStatus, _ := sample.Execute(&requestCtx)
			if httpStatus != http.StatusOK {
				panic("status not 200")
			}
		}
	})
}
