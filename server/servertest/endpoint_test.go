package servertest

import (
	"testing"
)

func Test_EndpointTest(t *testing.T) {
	t.Parallel()

	// Request should not be altered.
	requestBody := &createRequest{
		MyInt:      5,
		MyString:   "hello",
		ShouldFail: false,
	}

	createEndpoint := &create{
		Test:         "",
		RequestBody:  requestBody,
		ResponseBody: &createResponse{},
	}

	expectedHttpStatus := 201
	var expectedErr error
	expectedCreateEndpoint := &create{
		Test:        "",
		RequestBody: requestBody,
		ResponseBody: &createResponse{
			MyInt:    5,
			MyString: "hello",
		},
	}

	TestEndpoint(t, createEndpoint, expectedCreateEndpoint, expectedHttpStatus, expectedErr)
}
