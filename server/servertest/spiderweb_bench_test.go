package servertest

import (
	"net/http"
	"testing"

	"github.com/valyala/fasthttp"
)

func Benchmark_SpiderWeb_POST_latency(b *testing.B) {

	sample := routes()

	var req fasthttp.Request

	req.Header.SetMethod(http.MethodPost)
	req.Header.SetRequestURI("/sample")
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

	sample := routes()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var req fasthttp.Request

		req.Header.SetMethod(http.MethodPost)
		req.Header.SetRequestURI("/sample")
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
