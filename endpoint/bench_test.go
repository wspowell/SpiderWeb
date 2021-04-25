package endpoint

import (
	//"net/http"
	//_ "net/http/pprof"
	"net/http"
	"strings"
	"testing"

	"github.com/wspowell/context"
	"github.com/wspowell/log"
)

func Benchmark_Endpoint_Default_Success(b *testing.B) {
	endpointRunner := createTestEndpoint()

	// go func() {
	// 	_ = http.ListenAndServe("localhost:6060", nil)
	// }()

	ctx := context.Local()
	log.WithContext(ctx, log.NewConfig(log.LevelFatal))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {

		req, _ := http.NewRequest(http.MethodPost, "/resources/myid/5/true?id=me&num=13&flag=true", strings.NewReader(`{"my_string": "hello", "my_int": 5}`))

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")

		requester := NewHttpRequester("/resources/{id}/{num}/{flag}", req)

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

		req, _ := http.NewRequest(http.MethodPost, "/resources/myid/5/true?id=me&num=13&flag=true", strings.NewReader(`{"my_string": "hello", "my_int": 5}`))

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")

		requester := NewHttpRequester("/resources/{id}/{num}/{flag}", req)

		for pb.Next() {
			endpointRunner.Execute(ctx, requester)
		}
	})
}
