package profiling

// import (
// 	"io"

// 	opentracing "github.com/opentracing/opentracing-go"
// 	"github.com/uber/jaeger-client-go"
// 	config "github.com/uber/jaeger-client-go/config"
// 	"github.com/wspowell/context"
// 	"github.com/wspowell/log"
// )

// func StdOutTracer(appName string) (opentracing.Tracer, io.Closer) {
// 	cfg := &config.Configuration{
// 		ServiceName: appName,
// 		Sampler: &config.SamplerConfig{
// 			Type:  "const",
// 			Param: 1,
// 		},
// 		Reporter: &config.ReporterConfig{
// 			LogSpans: true,
// 		},
// 	}

// 	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
// 	if err != nil {
// 		log.Fatal(context.Background(), "failed to init Jaeger: %v", err)
// 	}

// 	return tracer, closer
// }
