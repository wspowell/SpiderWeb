package endpoint

import (
	//"net/http"
	//_ "net/http/pprof"
	"net/http"
	"strings"
	"testing"
)

func Benchmark_Endpoint_Default_Success(b *testing.B) {
	endpointRunner := createTestEndpoint()

	req, _ := http.NewRequest(http.MethodPost, "/resources/myid/5/true?id=me&num=13&flag=true", strings.NewReader(`{"my_string": "hello", "my_int": 5}`))

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	requester := NewHttpRequester("/resources/{id}/{num}/{flag}", req)

	// go func() {
	// 	_ = http.ListenAndServe("localhost:6060", nil)
	// }()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := newTestContext(requester)
			endpointRunner.Execute(ctx)
		}
	})
}

func Benchmark_Endpoint_Default_Error(b *testing.B) {
	endpointRunner := createTestEndpoint()

	req, _ := http.NewRequest(http.MethodPost, "/resources/myid/5/true?id=me&num=13&flag=true", strings.NewReader(`{"my_string": "hello", "my_int": 5}`))

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	requester := NewHttpRequester("/resources/{id}/{num}/{flag}", req)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := newTestContext(requester)
			endpointRunner.Execute(ctx)
		}
	})
}
