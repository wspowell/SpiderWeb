package endpoint

import (
	"testing"
)

func Benchmark_Endpoint_Default_Success(b *testing.B) {
	endpointRunner := createTestEndpoint()

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
