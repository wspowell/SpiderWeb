package endpoint

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wspowell/context"
)

type myRequestBodyModel struct {
	MyString   string `json:"myString"`
	MyInt      int    `json:"myInt"`
	ShouldFail bool   `json:"shouldFail"`
}

type myResponseBodyModel struct {
	OutputString string `json:"outputString"`
	OutputInt    int    `json:"outputInt"`
}

type testEndpointReqPtrResVal struct {
	Test         string
	RequestBody  *myRequestBodyModel `spiderweb:"request,mime=application/json,validate"`
	ResponseBody myResponseBodyModel `spiderweb:"response,mime=application/json,validate"`
}

func (self *testEndpointReqPtrResVal) Handle(ctx context.Context) (int, error) {
	return http.StatusOK, nil
}

const testValue = "test"

func Test_handlerTypeData_StructPtr_ReqPtr_ResVal(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	typeData := newHandlerTypeData(ctx, &testEndpointReqPtrResVal{})

	handlerAlloc := typeData.allocateHandler()
	if handler, ok := handlerAlloc.handler.(*testEndpointReqPtrResVal); !ok {
		t.Errorf("handler is not the right type")
	} else {
		if handler.Test != "" {
			t.Errorf("expected test value to be zero value")
		}
		handler.Test = testValue

		if handler.RequestBody.MyInt != 0 {
			t.Errorf("expected request body to be zero value")
		}
		handler.RequestBody.MyInt = 5

		if handler.ResponseBody.OutputInt != 0 {
			t.Errorf("expected response body to be zero value")
		}
		handler.ResponseBody.OutputInt = 5
	}

	handlerAlloc = typeData.allocateHandler()
	if handler, ok := handlerAlloc.handler.(*testEndpointReqPtrResVal); !ok {
		t.Errorf("handler is not the right type")
	} else {
		if handler.Test != "" {
			t.Errorf("expected test value to be zero value")
		}
		handler.Test = testValue

		if handler.RequestBody.MyInt != 0 {
			t.Errorf("expected request body to be zero value")
		}
		handler.RequestBody.MyInt = 5
		if handler.ResponseBody.OutputInt != 0 {
			t.Errorf("expected response body to be zero value")
		}
		handler.ResponseBody.OutputInt = 5
	}
}

type testEndpointNoReqResVal struct {
	Test         string
	ResponseBody myResponseBodyModel `spiderweb:"response,mime=application/json,validate"`
}

func (self *testEndpointNoReqResVal) Handle(ctx context.Context) (int, error) {
	return http.StatusOK, nil
}

// nolint:dupl // reason: too complex to refactor due to concrete type conversions
func Test_handlerTypeData_StructPtr_NoReq_ResVal(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	typeData := newHandlerTypeData(ctx, &testEndpointNoReqResVal{})

	handlerAlloc := typeData.allocateHandler()
	if handler, ok := handlerAlloc.handler.(*testEndpointNoReqResVal); !ok {
		t.Errorf("handler is not the right type")
	} else {
		if handler.Test != "" {
			t.Errorf("expected test value to be zero value")
		}
		handler.Test = testValue

		if handler.ResponseBody.OutputInt != 0 {
			t.Errorf("expected response body to be zero value")
		}
		handler.ResponseBody.OutputInt = 5
	}

	handlerAlloc = typeData.allocateHandler()
	if handler, ok := handlerAlloc.handler.(*testEndpointNoReqResVal); !ok {
		t.Errorf("handler is not the right type")
	} else {
		if handler.Test != "" {
			t.Errorf("expected test value to be zero value")
		}
		handler.Test = testValue

		if handler.ResponseBody.OutputInt != 0 {
			t.Errorf("expected response body to be zero value")
		}
		handler.ResponseBody.OutputInt = 5
	}
}

type testEndpointReqValResPtr struct {
	Test         string
	ResponseBody myResponseBodyModel `spiderweb:"response,mime=application/json,validate"`
}

func (self *testEndpointReqValResPtr) Handle(ctx context.Context) (int, error) {
	return http.StatusOK, nil
}

// nolint:dupl // reason: too complex to refactor due to concrete type conversions
func Test_handlerTypeData_StructPtr_ReqVal_ResPtr(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	typeData := newHandlerTypeData(ctx, &testEndpointReqValResPtr{})

	handlerAlloc := typeData.allocateHandler()
	if handler, ok := handlerAlloc.handler.(*testEndpointReqValResPtr); !ok {
		t.Errorf("handler is not the right type")
	} else {
		if handler.Test != "" {
			t.Errorf("expected test value to be zero value")
		}
		handler.Test = testValue

		if handler.ResponseBody.OutputInt != 0 {
			t.Errorf("expected response body to be zero value")
		}
		handler.ResponseBody.OutputInt = 5
	}

	handlerAlloc = typeData.allocateHandler()
	if handler, ok := handlerAlloc.handler.(*testEndpointReqValResPtr); !ok {
		t.Errorf("handler is not the right type")
	} else {
		if handler.Test != "" {
			t.Errorf("expected test value to be zero value")
		}
		handler.Test = testValue

		if handler.ResponseBody.OutputInt != 0 {
			t.Errorf("expected response body to be zero value")
		}
		handler.ResponseBody.OutputInt = 5
	}
}

func Test_handlerTypeData_no_etag(t *testing.T) {
	t.Parallel()

	type endpoint struct {
		ResponseBody *myResponseBodyModel `spiderweb:"response,mime=application/json"`
	}

	ctx := context.Background()

	typeData := newHandlerTypeData(ctx, &endpoint{})

	assert.False(t, typeData.eTagEnabled)
	assert.Equal(t, 0, typeData.maxAgeSeconds)
}

func Test_handlerTypeData_etag(t *testing.T) {
	t.Parallel()

	type endpoint struct {
		ResponseBody *myResponseBodyModel `spiderweb:"response,mime=application/json,etag"`
	}

	ctx := context.Background()

	typeData := newHandlerTypeData(ctx, &endpoint{})

	assert.True(t, typeData.eTagEnabled)
	assert.Equal(t, 0, typeData.maxAgeSeconds)
}

func Test_handlerTypeData_maxage(t *testing.T) {
	t.Parallel()

	type endpoint struct {
		ResponseBody *myResponseBodyModel `spiderweb:"response,mime=application/json,max-age=300"`
	}

	ctx := context.Background()

	typeData := newHandlerTypeData(ctx, &endpoint{})

	assert.False(t, typeData.eTagEnabled)
	assert.Equal(t, 300, typeData.maxAgeSeconds)
}

func Test_handlerTypeData_etag_maxage(t *testing.T) {
	t.Parallel()

	type endpoint struct {
		ResponseBody *myResponseBodyModel `spiderweb:"response,mime=application/json,etag,max-age=300"`
	}

	ctx := context.Background()

	typeData := newHandlerTypeData(ctx, &endpoint{})

	assert.True(t, typeData.eTagEnabled)
	assert.Equal(t, 300, typeData.maxAgeSeconds)
}
