package http_test

import (
	"fmt"
	"testing"

	"github.com/valyala/fasthttp"
	"github.com/wspowell/spiderweb/http"
	"github.com/wspowell/spiderweb/test"
)

func Benchmark_SpiderWeb_POST_latency(b *testing.B) {
	sample := test.Routes()

	var req fasthttp.Request

	req.Header.SetMethod(http.MethodPost)
	req.Header.SetRequestURI("/sample?for_bench=true")
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
			panic(fmt.Sprintf("status not 201: %v", httpStatus))
		}
	}
}

func Benchmark_SpiderWeb_POST_throughput(b *testing.B) {
	sample := test.Routes()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var req fasthttp.Request

		req.Header.SetMethod(http.MethodPost)
		req.Header.SetRequestURI("/sample?for_bench=true")
		req.Header.Set(fasthttp.HeaderHost, "localhost")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.SetBody([]byte(`{"output_string":"hello","output_int":5}`))

		requestCtx := fasthttp.RequestCtx{}
		requestCtx.Init(&req, nil, nil)

		for pb.Next() {
			httpStatus, _ := sample.Execute(&requestCtx)
			if httpStatus != http.StatusCreated {
				panic(fmt.Sprintf("status not 201: %v", httpStatus))
			}
		}
	})
}

func Benchmark_SpiderWeb_GET_latency(b *testing.B) {

	sample := test.Routes()

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

	sample := test.Routes()

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
