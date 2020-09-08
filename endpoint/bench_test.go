package endpoint

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"spiderweb/logging"
)

func Benchmark_Endpoint_Default_Success(b *testing.B) {
	endpointRunner := createTestEndpoint()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {

			logConfig := logging.NewConfig(logging.LevelFatal, map[string]interface{}{})
			req := httptest.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(`{"my_string": "hello", "my_int": 5}`)))
			ctx := NewContext(req, logging.NewLogger(logConfig))

			endpointRunner.Execute(ctx)

		}
	})
}

func Benchmark_Endpoint_Default_Error(b *testing.B) {
	endpointRunner := createTestEndpoint()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logConfig := logging.NewConfig(logging.LevelFatal, map[string]interface{}{})
			req := httptest.NewRequest(http.MethodGet, "/", bytes.NewBuffer([]byte(`{"my_string": "hello", "my_int": 5, "fail": true}`)))
			ctx := NewContext(req, logging.NewLogger(logConfig))

			endpointRunner.Execute(ctx)
		}
	})
}
