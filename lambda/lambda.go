package lambda

import (
	"github.com/wspowell/context"

	"github.com/wspowell/spiderweb/endpoint"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// FIXME: Should be able to handle any event, not just API Gateway.
// HandlerAPIGateway is an API Gateway Proxy Request handler function
type HandlerAPIGateway func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

type Lambda struct {
	matchedPath    string
	endpointConfig *endpoint.Config
	routeEndpoint  *endpoint.Endpoint
}

func New(endpointConfig *endpoint.Config, matchedPath string, handler endpoint.Handler) *Lambda {
	return &Lambda{
		matchedPath:    matchedPath,
		endpointConfig: endpointConfig,
		routeEndpoint:  endpoint.NewEndpoint(endpointConfig, handler),
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
		response := events.APIGatewayProxyResponse{}
		requester := NewApiGatewayRequester(self.matchedPath, &request)

		ctx, cancel := context.WithTimeout(ctx, self.endpointConfig.Timeout)
		go func() {
			<-ctx.Done()
			cancel()
		}()

		endpointCtx := endpoint.NewContext(ctx, requester)
		httpStatus, responseBody := routeEndpoint.Execute(endpointCtx)

		response.Body = string(responseBody)
		response.StatusCode = httpStatus
		response.Headers = requester.responseHeaders

		return response, nil
	}
}
