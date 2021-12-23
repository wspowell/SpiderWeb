package handler_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/wspowell/context"
	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/handler"
	"github.com/wspowell/spiderweb/request"
)

func Benchmark_Endpoint_Default_Success(b *testing.B) {
	handle := handler.NewHandle(testHandler{}).
		WithLogConfig(log.NewConfig().WithLevel(log.LevelFatal))

	ctx := context.Background()

	runner := handle.Runner()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/resources/myid/5/true?id=me&num=13&flag=true", strings.NewReader(`{"myString": "hello", "myInt": 5}`))
		if err != nil {
			panic(err)
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")

		requester, err := request.NewHttpRequester("/resources/{id}/{num}/{flag}", req)
		if err != nil {
			panic(err)
		}

		for pb.Next() {
			ctx := context.Localize(ctx)
			runner.Run(ctx, requester)
		}
	})
}

func Benchmark_Endpoint_Default_Error(b *testing.B) {
	handle := handler.NewHandle(testHandler{}).
		WithLogConfig(log.NewConfig().WithLevel(log.LevelFatal))

	ctx := context.Background()

	runner := handle.Runner()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/resources/myid/5/false?id=me&num=13&flag=true", strings.NewReader(`{"myString": "hello", "myInt": 5}`))
		if err != nil {
			panic(err)
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")

		requester, err := request.NewHttpRequester("/resources/{id}/{num}/{flag}", req)
		if err != nil {
			panic(err)
		}

		for pb.Next() {
			ctx := context.Localize(ctx)
			runner.Run(ctx, requester)
		}
	})
}
