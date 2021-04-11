package servertest

import (
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/wspowell/spiderweb/endpoint"
)

func handlerFuzzTest(t *testing.T, handler endpoint.Handler) {
	if doFuzz, exists := os.LookupEnv("FUZZ"); !exists || doFuzz != "true" {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			t.Fatalf("%+v\nhandler: %+v\n%+v", err, handler, string(debug.Stack()))
		}
	}()

	// The endpoint is never handed a struct with nil values so set nil chance to 0.
	f := fuzz.New().NilChance(0)

	endpointContext := endpoint.NewTestContext()
	for i := 0; i < 100; i++ {
		f.Fuzz(handler)
		handler.Handle(endpointContext)
	}
}

// TestEndpoint for business logic.
func TestEndpoint(t *testing.T, input endpoint.Handler, expected endpoint.Handler, expectedHttpStatus int, expectedError error) {
	endpointContext := endpoint.NewTestContext()

	httpStatus, err := input.Handle(endpointContext)

	if expectedError != err {
		t.Errorf("expected error %v, but got %v", expectedError, err)
	}

	if expectedHttpStatus != httpStatus {
		t.Errorf("expected http status %v, but got %v", expectedHttpStatus, httpStatus)
	}

	if !reflect.DeepEqual(expected, input) {
		t.Errorf("expected endpoint state does not equal actual:\n\texpected: \n%+v\n\tactual: \n%+v", deepPrintStruct(expected, 0), deepPrintStruct(input, 0))
	}

	handlerFuzzTest(t, input)
}

func deepPrintStruct(v interface{}, indentLevel int) string {
	var output string

	var indent string
	for i := 0; i < indentLevel; i++ {
		indent += "\t"
	}

	value := reflect.ValueOf(v)

	if value.Kind() == reflect.Ptr {
		output += "&"
		value = value.Elem()
	}
	output += value.Type().Name() + "{\n"
	for i := 0; i < value.NumField(); i++ {
		fieldValue := value.Field(i)
		fieldType := value.Type().Field(i)

		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.Elem().Kind() == reflect.Struct {
				output += fmt.Sprintf("%v\t%v: %v\n", indent, fieldType.Name, deepPrintStruct(fieldValue.Interface(), indentLevel+1))
			} else {
				output += fmt.Sprintf("%v\t%v: %v\n", indent, fieldType.Name, fieldValue.Elem().Interface())
			}
		} else {
			output += fmt.Sprintf("%v\t%v: %v\n", indent, fieldType.Name, fieldValue.Interface())
		}
	}

	output += indent + "}"

	return output
}
