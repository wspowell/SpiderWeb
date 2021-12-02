package restfultest

import (
	"bytes"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/mock"
	"github.com/valyala/fasthttp"

	"github.com/wspowell/spiderweb/server/restful"
)

var (
	// nolint:gochecknoglobals // reason: Remove the need for this by fixing runTest().
	// Tests alter the endpoint config for mocks, so these cannot run in parallel without locking.
	mutex = &sync.Mutex{}
)

type Mocker interface {
	AssertExpectations(t mock.TestingT) bool
}

type testCase struct {
	server *restful.Server
	name   string
}

func TestCase(server *restful.Server, name string) *testCase {
	return &testCase{
		server: server,
		name:   name,
	}
}

type requestTestCase struct {
	*testCase

	httpMethod      string
	path            string
	requestMimeType string
	requestBody     []byte
	headers         map[string]string
	queryParams     map[string]string
	pathParams      map[string]string
	resourceMocks   map[string]Mocker
}

// GivenRequest starts a request test case to be provided to TestRequest.
func (self *testCase) GivenRequest(httpMethod string, path string) *requestTestCase {
	return &requestTestCase{
		testCase:      self,
		httpMethod:    httpMethod,
		path:          path,
		headers:       map[string]string{},
		queryParams:   map[string]string{},
		pathParams:    map[string]string{},
		resourceMocks: map[string]Mocker{},
	}
}

func (self *requestTestCase) WithHeader(header string, value string) *requestTestCase {
	self.headers[header] = value

	return self
}

func (self *requestTestCase) WithQueryParam(param string, value string) *requestTestCase {
	self.queryParams[param] = value

	return self
}

func (self *requestTestCase) WithPathParam(param string, value string) *requestTestCase {
	self.pathParams[param] = value

	return self
}

// WithRequestBody sets a request body for the request test case.
// This is optional.
func (self *requestTestCase) WithRequestBody(mimeType string, requestBody []byte) *requestTestCase {
	self.requestMimeType = mimeType
	self.requestBody = requestBody

	return self
}

func (self *requestTestCase) WithResourceMock(resource string, resourceMock Mocker) *requestTestCase {
	self.resourceMocks[resource] = resourceMock

	return self
}

func (self *requestTestCase) ExpectResponse(httpStatus int) *responseTestCase {
	return &responseTestCase{
		testCase:   self.testCase,
		request:    self,
		httpStatus: httpStatus,
		headers:    map[string]string{},
	}
}

type responseTestCase struct {
	*testCase

	request *requestTestCase

	httpStatus       int
	headers          map[string]string
	responseMimeType string
	responseBody     []byte
	emptyBody        bool
}

func (self *responseTestCase) WithHeader(header string, value string) *responseTestCase {
	self.headers[header] = value

	return self
}

func (self *responseTestCase) WithEmptyBody() *responseTestCase {
	self.emptyBody = true

	return self
}

// Expect the response to match the given body.
func (self *responseTestCase) WithResponseBody(mimeType string, responseBody []byte) *responseTestCase {
	self.responseMimeType = mimeType
	self.responseBody = responseBody

	return self
}

func (self *responseTestCase) Run(t *testing.T) {
	t.Helper()
	t.Run(self.name, func(t *testing.T) {
		self.runTest(t)
	})
}

func (self *responseTestCase) RunParallel(t *testing.T) {
	t.Helper()
	t.Run(self.name, func(t *testing.T) {
		t.Parallel()
		self.runTest(t)
	})
}

func (self *responseTestCase) runTest(t *testing.T) {
	t.Helper()

	// FIXME: Tests alter the endpoint config for mocks, so these cannot run in parallel without locking.
	mutex.Lock()
	defer mutex.Unlock()

	copyRequestBody := make([]byte, len(self.request.requestBody))
	copy(copyRequestBody, self.request.requestBody)

	copyResponseBody := make([]byte, len(self.responseBody))
	copy(copyResponseBody, self.responseBody)

	copyTestCase := responseTestCase{
		request: &requestTestCase{
			httpMethod:      self.request.httpMethod,
			path:            self.request.path,
			requestMimeType: self.request.requestMimeType,
			requestBody:     copyRequestBody,
			headers:         self.request.headers,
			queryParams:     self.request.queryParams,
			pathParams:      self.request.pathParams,
			resourceMocks:   self.request.resourceMocks,
		},

		httpStatus:       self.httpStatus,
		responseMimeType: self.responseMimeType,
		responseBody:     copyResponseBody,
		headers:          self.headers,
		emptyBody:        self.emptyBody,
	}

	url := copyTestCase.request.path
	for param, value := range copyTestCase.request.pathParams {
		url = strings.Replace(url, "{"+param+"}", value, 1)
	}

	var req fasthttp.Request

	req.Header.SetMethod(copyTestCase.request.httpMethod)
	req.Header.SetRequestURI(url)
	req.Header.Set(fasthttp.HeaderHost, "localhost")

	if copyTestCase.request.requestMimeType == "" && copyTestCase.responseMimeType == "" {
		copyTestCase.request.requestMimeType = "application/json"
		copyTestCase.responseMimeType = "application/json"
	} else if copyTestCase.responseMimeType == "" {
		copyTestCase.responseMimeType = copyTestCase.request.requestMimeType
	}
	req.Header.Set("Content-Type", copyTestCase.request.requestMimeType)
	req.Header.Set("Accept", copyTestCase.responseMimeType)

	for header, value := range copyTestCase.request.headers {
		req.Header.Set(header, value)
	}

	req.SetBody(copyTestCase.request.requestBody)

	requestCtx := fasthttp.RequestCtx{}
	requestCtx.Init(&req, nil, nil)

	// Setup mock calls.
	endpoint := self.server.Endpoint(copyTestCase.request.httpMethod, copyTestCase.request.path)
	originalResources := map[string]interface{}{}
	if endpoint != nil {
		for name, resource := range endpoint.Config.Resources {
			originalResources[name] = resource
			if resourceMock, ok := copyTestCase.request.resourceMocks[name]; ok {
				endpoint.Config.Resources[name] = resourceMock
			} else {
				// Do not call resources. Must be mocked.
				endpoint.Config.Resources[name] = nil
			}
		}
	}

	actualHttpStatus, actualResponseBody := self.server.Execute(&requestCtx)

	if endpoint != nil {
		// Put the resources back.
		for name, originalResource := range originalResources {
			if resourceMock, ok := copyTestCase.request.resourceMocks[name]; ok {
				resourceMock.AssertExpectations(t)
			}
			endpoint.Config.Resources[name] = originalResource
		}
	}

	for header, value := range copyTestCase.headers {
		actualHeaderValue := requestCtx.Response.Header.Peek(header)
		if !bytes.Equal(actualHeaderValue, []byte(value)) {
			t.Errorf("expected header %v = %v , but got %v = %v", header, value, header, actualHeaderValue)
		}
	}

	if copyTestCase.httpStatus != actualHttpStatus {
		t.Errorf("expected http status %v, but got %v", copyTestCase.httpStatus, actualHttpStatus)
	}

	if copyTestCase.emptyBody {
		if !bytes.Equal(nil, actualResponseBody) {
			t.Errorf("expected empty response body, but got '%v'", string(actualResponseBody))
		}
	} else {
		if requestCtx.Response.Header.ContentType() == nil {
			t.Errorf("response is missing header Content-Type")
		} else if !bytes.Equal(requestCtx.Response.Header.ContentType(), []byte(copyTestCase.responseMimeType)) {
			t.Errorf("expected response mime type '%v', but got '%v'", copyTestCase.responseMimeType, requestCtx.Response.Header.ContentType())
		}

		if !bytes.Equal(copyTestCase.responseBody, actualResponseBody) {
			t.Errorf("expected response body '%v', but got '%v'", string(copyTestCase.responseBody), string(actualResponseBody))
		}
	}

	requestFuzzTest(t, self.server, self.request.httpMethod, self.request.path)
}

func requestFuzzTest(t *testing.T, server *restful.Server, httpMethod string, path string) {
	t.Helper()

	if doFuzz, exists := os.LookupEnv("FUZZ"); !exists || doFuzz != "true" {
		return
	}

	var requestBody []byte
	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("%+v\route: %v %v\nrequest body: %+v\n%+v", err, httpMethod, path, string(requestBody), string(debug.Stack()))
		}
	}()

	f := fuzz.New()

	for i := 0; i < 100; i++ {
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
}
