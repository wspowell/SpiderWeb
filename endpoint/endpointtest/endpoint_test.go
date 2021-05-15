package endpointtest_test

import (
	"testing"

	"github.com/wspowell/spiderweb/endpoint/endpointtest"
	"github.com/wspowell/spiderweb/test"
)

func Test_EndpointTest(t *testing.T) {
	t.Parallel()

	// Request should not be altered.
	requestBody := &test.CreateRequest{
		MyInt:      5,
		MyString:   "hello",
		ShouldFail: false,
	}

	createEndpoint := &test.Create{
		Test:         "",
		RequestBody:  requestBody,
		ResponseBody: &test.CreateResponse{},
	}

	expectedHttpStatus := 201
	var expectedErr error
	expectedCreateEndpoint := &test.Create{
		Test:        "",
		RequestBody: requestBody,
		ResponseBody: &test.CreateResponse{
			MyInt:    5,
			MyString: "hello",
		},
	}

	endpointtest.TestEndpoint(t, createEndpoint, expectedCreateEndpoint, expectedHttpStatus, expectedErr)
}
