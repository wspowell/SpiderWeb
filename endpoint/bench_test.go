package endpoint

import (
	//"net/http"
	//_ "net/http/pprof"
	"testing"
)

func Benchmark_Endpoint_Default_Success(b *testing.B) {
	endpointRunner := createTestEndpoint()

	// go func() {
	// 	_ = http.ListenAndServe("localhost:6060", nil)
	// }()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := newTestContext()
			endpointRunner.Execute(ctx)
		}
	})
}

func Benchmark_Endpoint_Default_Error(b *testing.B) {
	endpointRunner := createTestEndpoint()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := newTestContext()
			endpointRunner.Execute(ctx)
		}
	})
}
