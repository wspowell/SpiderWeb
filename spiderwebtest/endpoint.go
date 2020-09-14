package spiderwebtest

import (
	"fmt"
	"reflect"
	"testing"

	"spiderweb/endpoint"
)

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
