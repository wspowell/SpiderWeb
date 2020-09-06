package restful

import (
	"net/http"

	"spiderweb/local"
	"spiderweb/logging"
	"spiderweb/profiling"
)

// Context for a routed request. This contains everything an endpoint needs
// to make business logic decisions. This context is created by the router
// middleware in order to provide a well defined layer between the endpoint
// and the HTTP framework.
type Context struct {
	*local.Localized

	request *http.Request
	// route is the designed routed path, not the incoming URL path.
	// Ex: /resource/:id
	route string

	// pathParameters matched in the route.
	// Ex: /resource/:id -> { "id": "my_resource_id" }
	pathParameters map[string]string
}

func NewContext(request *http.Request, route string, pathParameters map[string]string) *Context {
	return &Context{
		Localized:      local.NewLocalized(request.Context()),
		request:        request,
		route:          route,
		pathParameters: pathParameters,
	}
}

type LoggerContext struct {
	local.Localizer

	Logger *logging.Logger
}

func NewLoggerContext(ctx local.Localizer, logConfig logging.Configurer) *LoggerContext {
	return &LoggerContext{
		Localizer: ctx,
		Logger:    logging.NewLogger(logConfig),
	}
}

type LoggerRoundTripper struct {
	RoundTripper

	LogConfig logging.Configurer
}

func (self *LoggerRoundTripper) RoundTrip(handler local.Localizer) ([]byte, int) {
	// Wrap the handler with a logger.
	handler = NewLoggerContext(handler, self.LogConfig)

	return self.RoundTripper.RoundTrip(handler)
}

type ProfileContext struct {
	local.Localizer

	Profiler *profiling.OpenTracingProfiler
}

func NewProfileContext(ctx local.Localizer, operationName string) *ProfileContext {
	transaction, spanCtx := profiling.NewOpenTracingProfiler(ctx, operationName)

	return &ProfileContext{
		Localizer: local.NewLocalized(spanCtx), // FIXME: This overwrites the passed in localizer
		Profiler:  transaction,
	}
}

type ProfilerRoundTripper struct {
	RoundTripper

	OperationName string
}

func (self *ProfilerRoundTripper) RoundTrip(handler local.Localizer) ([]byte, int) {
	// Wrap the handler with a logger.
	profiler := NewProfileContext(handler, self.OperationName)
	defer profiler.Profiler.Finish()

	handler = profiler

	return self.RoundTripper.RoundTrip(handler)
}
