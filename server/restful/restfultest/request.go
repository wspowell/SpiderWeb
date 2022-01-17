package restfultest

import (
	"bytes"
	"os"
	"strings"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/valyala/fasthttp"

	"github.com/wspowell/errors"
	"github.com/wspowell/spiderweb/httpheader"
	"github.com/wspowell/spiderweb/server/restful"
)

type testCase struct {
	server *restful.Server
	name   string

	httpMethod     string
	path           string
	requestBody    []byte
	requestHeaders map[string]string
	queryParams    map[string]string
	pathParams     map[string]string

	httpStatus      int
	responseHeaders map[string]string
	responseBody    []byte
	emptyBody       bool
}

type caseBuilder struct {
	*testCase
}

func Server(server *restful.Server) *caseBuilder {
	return &caseBuilder{
		testCase: &testCase{
			server:          server,
			emptyBody:       true,
			requestHeaders:  map[string]string{},
			queryParams:     map[string]string{},
			pathParams:      map[string]string{},
			responseHeaders: map[string]string{},
		},
	}
}

func (self *caseBuilder) Case(name string) *routeBuilder {
	self.testCase.name = name
	return &routeBuilder{self.testCase}
}

type routeBuilder struct {
	*testCase
}

// ForRoute starts a request test case to be provided to TestRequest.
func (self *routeBuilder) ForRoute(httpMethod string, path string) *requestBuilder {
	self.testCase.httpMethod = httpMethod
	self.testCase.path = path
	return &requestBuilder{self.testCase}
}

type requestBuilder struct {
	*testCase
}

func (self *requestBuilder) WithHeader(header string, value string) *requestBuilder {
	self.testCase.requestHeaders[header] = value

	return self
}

func (self *requestBuilder) WithQueryParam(param string, value string) *requestBuilder {
	self.testCase.queryParams[param] = value

	return self
}

func (self *requestBuilder) WithPathParam(param string, value string) *requestBuilder {
	self.testCase.pathParams[param] = value

	return self
}

// WithRequestBody sets a request body for the request test case.
// This is optional.
func (self *requestBuilder) WithRequestBody(requestBody []byte) *requestBuilder {
	self.requestBody = requestBody

	return self
}

type responseBuilder struct {
	*testCase
}

func (self *requestBuilder) ExpectStatusCode(httpStatus int) *responseBuilder {
	self.testCase.httpStatus = httpStatus
	return &responseBuilder{self.testCase}
}

func (self *responseBuilder) WithHeader(header string, value string) *responseBuilder {
	self.testCase.responseHeaders[header] = value

	return self
}

// Expect the response to match the given body.
func (self *responseBuilder) WithResponseBody(responseBody []byte) *responseBuilder {
	self.emptyBody = false
	self.responseBody = responseBody

	return self
}

func (self *responseBuilder) Run(t *testing.T) {
	t.Helper()
	t.Run(self.name, func(t *testing.T) {
		self.runTest(t)
	})
}

func (self *responseBuilder) RunParallel(t *testing.T) {
	t.Helper()
	t.Run(self.name, func(t *testing.T) {
		t.Parallel()
		self.runTest(t)
	})
}

func (self *testCase) runTest(t *testing.T) {
	t.Helper()

	url := self.path
	for param, value := range self.pathParams {
		url = strings.Replace(url, "{"+param+"}", value, 1)
	}

	var req fasthttp.Request

	req.Header.SetMethod(self.httpMethod)
	req.Header.SetRequestURI(url)
	req.Header.Set(fasthttp.HeaderHost, "localhost")

	for header, value := range self.requestHeaders {
		req.Header.Set(header, value)
	}

	req.SetBody(self.requestBody)

	requestCtx := fasthttp.RequestCtx{}
	requestCtx.Init(&req, nil, nil)

	actualHttpStatus, actualResponseBody := self.server.Execute(&requestCtx)

	for responseHeader, value := range self.responseHeaders {
		actualHeaderValue := requestCtx.Response.Header.Peek(responseHeader)
		if !bytes.Equal(actualHeaderValue, []byte(value)) {
			t.Errorf("expected response header %v = %v , but got %v = %v", responseHeader, value, responseHeader, actualHeaderValue)
		}
	}

	if self.httpStatus != actualHttpStatus {
		t.Errorf("expected http status %v, but got %v", self.httpStatus, actualHttpStatus)
	}

	if self.emptyBody {
		if !bytes.Equal(nil, actualResponseBody) {
			t.Errorf("expected empty response body, but got '%s'", actualResponseBody)
		}
	} else {
		if requestCtx.Response.Header.ContentType() == nil {
			t.Errorf("response is missing header Content-Type")
		} else if !bytes.Equal([]byte(self.responseHeaders[httpheader.ContentType]), requestCtx.Response.Header.ContentType()) {
			t.Errorf("expected response mime type '%v', but got '%v'", self.responseHeaders[httpheader.ContentType], requestCtx.Response.Header.ContentType())
		}

		if !bytes.Equal(self.responseBody, actualResponseBody) {
			t.Errorf("expected response body '%v', but got '%v'", string(self.responseBody), string(actualResponseBody))
		}
	}

	requestFuzzTest(t, self.server, self.httpMethod, self.path)
}

func requestFuzzTest(t *testing.T, server *restful.Server, httpMethod string, path string) {
	t.Helper()

	if doFuzz, exists := os.LookupEnv("FUZZ"); !exists || doFuzz != "true" {
		return
	}

	var requestBody []byte

	err := errors.Catch(func() {
		for i := 0; i < 10000; i++ {
			// TODO: Make this work better.
			f := fuzz.New().
				NilChance(0.1).
				MaxDepth(20).
				NumElements(0, 100)

			f.Fuzz(&requestBody)

			var req fasthttp.Request

			req.Header.SetMethod(httpMethod)
			req.Header.SetRequestURI(path)
			req.Header.Set(fasthttp.HeaderHost, "localhost")
			req.SetBody(requestBody)

			requestCtx := fasthttp.RequestCtx{}
			requestCtx.Init(&req, nil, nil)

			server.Execute(&requestCtx)
		}
	})

	if err != nil {
		t.Errorf("fuzz test caused panice in route(%v %v): %+v", httpMethod, path, err)
	}
}
