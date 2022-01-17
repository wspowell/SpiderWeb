package lambda

import (
	"strings"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/wspowell/spiderweb/httpheader"
	"github.com/wspowell/spiderweb/httptrip"
)

var (
	bytesPool = sync.Pool{
		New: func() interface{} {
			// The Pool's New function should generally only return pointer
			// types, since a pointer can be put into the return interface
			// value without an allocation:
			return []byte{}
		},
	}
)

var _ httptrip.RoundTripper = (*ApiGatewayRequester)(nil)

type ApiGatewayRequester struct {
	matchedPath string
	request     *events.APIGatewayProxyRequest
	bodyBytes   []byte

	response          events.APIGatewayProxyResponse
	responseHeaders   map[string]string
	statusCode        int
	responseBodyBytes []byte
}

func NewApiGatewayRequester(matchedPath string, request *events.APIGatewayProxyRequest) *ApiGatewayRequester {
	var bodyBytes []byte
	if request.Body != "" {
		bodyBytes = []byte(request.Body)
	}

	responseBodyBytes := bytesPool.Get().([]byte)
	responseBodyBytes = responseBodyBytes[:0]

	return &ApiGatewayRequester{
		matchedPath:       matchedPath,
		request:           request,
		bodyBytes:         bodyBytes,
		response:          events.APIGatewayProxyResponse{},
		responseHeaders:   map[string]string{},
		responseBodyBytes: responseBodyBytes,
	}
}

func (self *ApiGatewayRequester) Close() {
	bytesPool.Put(self.responseBodyBytes)
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
	return []byte(self.request.Headers[httpheader.ContentType])
}

func (self *ApiGatewayRequester) Accept() []byte {
	return []byte(self.request.Headers[httpheader.Accept])
}

func (self *ApiGatewayRequester) PeekRequestHeader(key string) []byte {
	if value, exists := self.request.Headers[key]; exists {
		return []byte(value)
	}

	return nil
}

func (self *ApiGatewayRequester) VisitRequestHeaders(f func(key []byte, value []byte)) {
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
	paramVariable := "{" + param + "}"

	for index, value := range pathParts {
		if value == paramVariable {
			return urlParts[index], true
		}
	}

	return "", false
}

func (self *ApiGatewayRequester) QueryParam(param string) (string, bool) {
	value, exists := self.request.QueryStringParameters[param]

	return value, exists
}

func (self *ApiGatewayRequester) RequestBody() []byte {
	return self.bodyBytes
}

func (self *ApiGatewayRequester) ResponseBody() []byte {
	return self.responseBodyBytes
}

func (self *ApiGatewayRequester) StatusCode() int {
	return self.statusCode
}

func (self *ApiGatewayRequester) SetStatusCode(statusCode int) {
	self.statusCode = statusCode
}

func (self *ApiGatewayRequester) SetResponseHeader(header string, value string) {
	self.responseHeaders[string(header)] = string(value)
}

func (self *ApiGatewayRequester) SetResponseBody(body []byte) {
	self.responseBodyBytes = body
}

func (self *ApiGatewayRequester) SetResponseContentType(contentType string) {
	self.responseHeaders[httpheader.ContentType] = string(contentType)
}

func (self *ApiGatewayRequester) ResponseContentType() []byte {
	return []byte(self.responseHeaders[httpheader.ContentType])
}

func (self *ApiGatewayRequester) PeekResponseHeader(header string) []byte {
	value, exists := self.responseHeaders[header]
	if exists {
		return []byte(value)
	}
	return nil
}

func (self *ApiGatewayRequester) VisitResponseHeaders(f func(header []byte, value []byte)) {
	for header, value := range self.responseHeaders {
		f([]byte(header), []byte(value))
	}
}

func (self *ApiGatewayRequester) WriteResponse() {
	self.response.Body = string(self.responseBodyBytes)
	self.response.StatusCode = self.statusCode
	self.response.Headers = self.responseHeaders
}
