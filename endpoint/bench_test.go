package endpoint_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/wspowell/context"
	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/endpoint"
)

func Benchmark_Endpoint_Default_Success(b *testing.B) {
	endpointRunner := createTestEndpoint()

	ctx := context.Local()
	log.WithContext(ctx, log.NewConfig(log.LevelFatal))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/resources/myid/5/true?id=me&num=13&flag=true", strings.NewReader(`{"myString": "hello", "myInt": 5}`))
		if err != nil {
			panic(err)
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")

		requester, err := endpoint.NewHttpRequester("/resources/{id}/{num}/{flag}", req)
		if err != nil {
			panic(err)
		}

		for pb.Next() {
			endpointRunner.Execute(ctx, requester)
		}
	})
}

func Benchmark_Endpoint_Default_Error(b *testing.B) {
	endpointRunner := createTestEndpoint()

	ctx := context.Local()
	log.WithContext(ctx, log.NewConfig(log.LevelFatal))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/resources/myid/5/true?id=me&num=13&flag=true", strings.NewReader(`{"myString": "hello", "myInt": 5}`))
		if err != nil {
			panic(err)
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")

		requester, err := endpoint.NewHttpRequester("/resources/{id}/{num}/{flag}", req)
		if err != nil {
			panic(err)
		}

		for pb.Next() {
			endpointRunner.Execute(ctx, requester)
		}
	})
}
