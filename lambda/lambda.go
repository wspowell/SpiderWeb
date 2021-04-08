package lambda

import (
	"context"
	"net/http"

	"github.com/wspowell/logging"
	"github.com/wspowell/spiderweb/endpoint"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/valyala/fasthttp"
)

// FIXME: Should be able to handle any event, not just API Gateway.
// HandlerAPIGateway is an API Gateway Proxy Request handler function
type HandlerAPIGateway func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

type Lambda struct {
	lambdaContext  context.Context
	endpointConfig *endpoint.Config
	routeEndpoint  *endpoint.Endpoint
}

func NewLambda(endpointConfig *endpoint.Config, handler endpoint.Handler) *Lambda {
	return &Lambda{
		lambdaContext:  context.Background(),
		endpointConfig: endpointConfig,
		routeEndpoint:  endpoint.NewEndpoint(endpointConfig, handler),
	}
}

func (self *Lambda) Start() {
	wrappedHandler := self.wrapLambdaHandler(self.routeEndpoint)

	lambda.Start(wrappedHandler)
}

// func (self *Lambda) Execute() (int, []byte) {
// 	self.router.Handler(fasthttpCtx)
// 	return fasthttpCtx.Response.StatusCode(), fasthttpCtx.Response.Body()
// }

// func (self *Lambda) Invoke(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

// }

func (self *Lambda) wrapLambdaHandler(routeEndpoint *endpoint.Endpoint) HandlerAPIGateway {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		response := events.APIGatewayProxyResponse{}

		// FIXME: This should be able to execute without fasthttp.
		var req fasthttp.Request

		req.Header.SetMethod(request.HTTPMethod)
		req.Header.SetRequestURI(request.Path)
		for header, value := range request.Headers {
			req.Header.Set(header, value)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.SetBody([]byte(request.Body))

		requestCtx := &fasthttp.RequestCtx{}
		requestCtx.Init(&req, nil, nil)

		fasthttp.TimeoutWithCodeHandler(func(fasthttpCtx *fasthttp.RequestCtx) {
			// Every invocation of an endpoint is guaranteed to get its own logger instance.
			var logger logging.Logger = logging.NewLog(self.endpointConfig.LogConfig)

			logger.Tag("request_id", fasthttpCtx.ID())
			logger.Tag("route", request.HTTPMethod+" "+request.Resource)

			// Note: The endpoint context must receive the same timeout as the fasthttp.TimeoutWithCodeHandler or this will cause unexpected behavior.
			ctx := endpoint.NewContext(self.lambdaContext, fasthttpCtx, logger, self.endpointConfig.Timeout)
			httpStatus, responseBody := routeEndpoint.Execute(ctx)

			fasthttpCtx.SetStatusCode(httpStatus)
			fasthttpCtx.SetBody(responseBody)

			headers := map[string]string{}
			fasthttpCtx.Response.Header.VisitAll(func(key []byte, value []byte) {
				headers[string(key)] = string(value)
			})

			response.Body = string(responseBody)
			response.StatusCode = httpStatus
			response.Headers = headers

			// Set the Connection header to "close".
			// Closes the connection after this function returns.
			fasthttpCtx.Response.SetConnectionClose()
		}, self.endpointConfig.Timeout, "", http.StatusRequestTimeout)(requestCtx)

		return response, nil
	}
}
