package lambda

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/wspowell/context"
	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/endpoint"
	"github.com/wspowell/spiderweb/server/route"
)

// FIXME: Should be able to handle any event, not just API Gateway.
// HandlerAPIGateway is an API Gateway Proxy Request handler function
type HandlerAPIGateway func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

type Lambda struct {
	endpointConfig *endpoint.Config
	matchedPath    string
	routeEndpoint  *endpoint.Endpoint
}

func New(endpointConfig *endpoint.Config, routeDefinition route.Route) *Lambda {
	ctx := context.Local()
	log.WithContext(ctx, endpointConfig.LogConfig)

	return &Lambda{
		endpointConfig: endpointConfig,
		matchedPath:    routeDefinition.Path,
		routeEndpoint:  endpoint.NewEndpoint(ctx, endpointConfig, routeDefinition.Handler),
	}
}

func (self *Lambda) Start() {
	wrappedHandler := self.wrapLambdaHandler(self.routeEndpoint)

	lambda.Start(wrappedHandler)
}

// FIXME: Should be able to execute a lambda, especially for testing.
// func (self *Lambda) Execute() (int, []byte) {
// 	self.router.Handler(fasthttpCtx)
// 	return fasthttpCtx.Response.StatusCode(), fasthttpCtx.Response.Body()
// }

// func (self *Lambda) Invoke(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

// }

func (self *Lambda) wrapLambdaHandler(routeEndpoint *endpoint.Endpoint) HandlerAPIGateway {
	return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, routeEndpoint.Config.Tracer, request.HTTPMethod+" "+self.matchedPath)
		defer span.Finish()

		response := events.APIGatewayProxyResponse{}
		requester := NewApiGatewayRequester(self.matchedPath, &request)

		ctx, cancel := context.WithTimeout(ctx, self.endpointConfig.Timeout)
		go func() {
			<-ctx.Done()
			cancel()
		}()

		httpStatus, responseBody := routeEndpoint.Execute(ctx, requester)

		response.Body = string(responseBody)
		response.StatusCode = httpStatus
		response.Headers = requester.responseHeaders

		return response, nil
	}
}
