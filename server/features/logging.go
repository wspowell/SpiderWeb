package features

// import (
// 	"spiderweb/logging"
// 	"spiderweb/server/endpoint"
// )

// type logger interface {
// 	SetLogger(loggerConfig logging.Configurer)
// }

// func Logging(loggerConfig logging.Configurer) FeatureFunc {
// 	return func(handler endpoint.Handler) {
// 		if endpointLogger, ok := handler.(logger); ok {
// 			endpointLogger.SetLogger(loggerConfig)
// 		}
// 	}
// }
