package spiderweb

import (
	"fmt"

	"spiderweb/endpoint"
	"spiderweb/logging"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

// Config top level options.
// These options can be altered per endpoint, if desired.
type Config struct {
	EndpointConfig endpoint.Config
	ServerHost     string
	ServerPort     int
}

// Clone Config since endpoint.Config also Clones.
func (self Config) Clone() Config {
	return Config{
		EndpointConfig: self.EndpointConfig.Clone(),
		ServerHost:     self.ServerHost,
		ServerPort:     self.ServerPort,
	}
}

type framework struct {
	config Config
	logger *logging.Logger
	router *router.Router

	endpointBuilders []*endpointBuilder
}

func New(config Config) *framework {
	return &framework{
		config: config,
		logger: logging.NewLogger(config.EndpointConfig.LogConfig),
		router: router.New(),
	}
}

func (self *framework) Handle(httpMethod string, path string, handler endpoint.Handler) *endpointBuilder {
	builder := &endpointBuilder{
		httpMethod: httpMethod,
		path:       path,
		builder:    endpoint.NewEndpoint(self.config.EndpointConfig.Clone(), handler),
	}

	self.endpointBuilders = append(self.endpointBuilders, builder)

	return builder
}

func (self *framework) Run() {
	self.setupEndpoints()
	self.listenForever()
}

func (self *framework) setupEndpoints() {
	for _, endpointBuilder := range self.endpointBuilders {
		httpMethod := endpointBuilder.httpMethod
		path := endpointBuilder.path
		handler := wrapFasthttpHandler(endpointBuilder)

		self.router.Handle(httpMethod, path, handler)
	}
}

func (self *framework) listenForever() {
	for key, list := range self.router.List() {
		self.logger.Debug("%v", key)
		for _, item := range list {
			self.logger.Debug("  %v", item)
		}
	}

	self.logger.Debug("listening...")

	listenAddress := fmt.Sprintf("%s:%d", self.config.ServerHost, self.config.ServerPort)
	if err := fasthttp.ListenAndServe(listenAddress, self.router.Handler); err != nil {
		self.logger.Fatal("listener failed: %v", err)
	}
}

type endpointBuilder struct {
	httpMethod string
	path       string
	builder    *endpoint.Endpoint
}

func newEndpointBuilder(config Config, handler endpoint.Handler) *endpointBuilder {
	return &endpointBuilder{
		builder: endpoint.NewEndpoint(config.EndpointConfig.Clone(), handler),
	}
}

func (self *endpointBuilder) WithErrorHandling(errorHandler endpoint.ErrorHandler) *endpointBuilder {
	self.builder.Config.ErrorHandler = errorHandler
	return self
}

func (self *endpointBuilder) WithAuth(auther endpoint.Auther) *endpointBuilder {
	self.builder.Config.Auther = auther
	return self
}

func (self *endpointBuilder) WithRequestValidation(requestValidator endpoint.RequestValidator) *endpointBuilder {
	self.builder.Config.RequestValidator = requestValidator
	return self
}

func (self *endpointBuilder) WithResponseValidation(responseValidator endpoint.ResponseValidator) *endpointBuilder {
	self.builder.Config.ResponseValidator = responseValidator
	return self
}

func (self *endpointBuilder) WithMimeType(mimeType string, handler endpoint.MimeTypeHandler) *endpointBuilder {
	self.builder.Config.MimeTypeHandlers[mimeType] = handler
	return self
}

func wrapFasthttpHandler(endpointRunner *endpointBuilder) fasthttp.RequestHandler {
	return func(fasthttpCtx *fasthttp.RequestCtx) {
		logger := logging.NewLogger(endpointRunner.builder.Config.LogConfig)
		ctx := endpoint.NewContext(fasthttpCtx, logger)
		httpStatus, responseBody := endpointRunner.builder.Execute(ctx)

		fasthttpCtx.SetStatusCode(httpStatus)
		fasthttpCtx.SetBody(responseBody)
	}
}
