package features

// import (
// 	"context"

// 	"spiderweb/profiling"
// 	"spiderweb/server/endpoint"
// )

// type profiler interface {
// 	New(ctx context.Context, operationName string) (profiling.Profiler, context.Context)
// }

// func Profiling(ctx context.Context, operationName string) FeatureFunc {
// 	return func(handler endpoint.Handler) {
// 		if endpointProfiler, ok := handler.(profiler); ok {
// 			endpointProfiler.New(ctx, operationName)
// 		}
// 	}
// }
