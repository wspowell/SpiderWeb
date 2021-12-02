package lambda

import (
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

type ApiGatewayRequester struct {
	matchedPath string
	request     *events.APIGatewayProxyRequest
	bodyBytes   []byte

	responseHeaders map[string]string
}

func NewApiGatewayRequester(matchedPath string, request *events.APIGatewayProxyRequest) *ApiGatewayRequester {
	var bodyBytes []byte
	if request.Body != "" {
		bodyBytes = []byte(request.Body)
	}

	return &ApiGatewayRequester{
		matchedPath:     matchedPath,
		request:         request,
		bodyBytes:       bodyBytes,
		responseHeaders: map[string]string{},
	}
}

func (self *ApiGatewayRequester) RequestId() string {
	return self.request.RequestContext.RequestID
}

func (self *ApiGatewayRequester) Method() []byte {
	return []byte(self.request.HTTPMethod)
}

func (self *ApiGatewayRequester) Path() []byte {
	return []byte(self.request.Path)
}

func (self *ApiGatewayRequester) ContentType() []byte {
	return []byte(self.request.Headers["Content-Type"])
}

func (self *ApiGatewayRequester) Accept() []byte {
	return []byte(self.request.Headers["Accept"])
}

func (self *ApiGatewayRequester) PeekHeader(key string) []byte {
	if value, exists := self.request.Headers[key]; exists {
		return []byte(value)
	}

	return nil
}

func (self *ApiGatewayRequester) VisitHeaders(f func(key []byte, value []byte)) {
	for header, value := range self.request.Headers {
		f([]byte(header), []byte(value))
	}
}

func (self *ApiGatewayRequester) MatchedPath() string {
	return self.matchedPath
}

func (self *ApiGatewayRequester) PathParam(param string) (string, bool) {
	urlParts := strings.Split(self.request.Path, "/")
	pathParts := strings.Split(self.matchedPath, "/")

	for index, value := range pathParts {
		if value == "{"+param+"}" {
			return urlParts[index], true
		}
	}

	return "", false
}

func (self *ApiGatewayRequester) QueryParam(param string) ([]byte, bool) {
	value, exists := self.request.QueryStringParameters[param]

	return []byte(value), exists
}

func (self *ApiGatewayRequester) RequestBody() []byte {
	return self.bodyBytes
}

func (self *ApiGatewayRequester) SetResponseHeader(header string, value string) {
	self.responseHeaders[header] = value
}

func (self *ApiGatewayRequester) SetResponseContentType(contentType string) {
	self.responseHeaders["Content-Type"] = contentType
}

func (self *ApiGatewayRequester) ResponseContentType() string {
	return self.responseHeaders["Content-Type"]
}

func (self *ApiGatewayRequester) ResponseHeaders() map[string]string {
	return self.responseHeaders
}
