package endpoint

import (
	"reflect"
	"strings"
)

// This file contains all the reflection that is not nice to look at.

func newStructFromHandler(handler Handler, isHandlerPtr bool, fieldNum int) interface{} {
	var handlerValue reflect.Value
	if isHandlerPtr {
		handlerValue = reflect.ValueOf(handler).Elem()
	} else {
		handlerValue = reflect.ValueOf(handler)
	}

	if !handlerValue.Field(fieldNum).IsValid() {
		return nil
	}

	newValue := handlerValue.Field(fieldNum)
	return newValue.Addr().Interface()
}

// handlerTypeData cached so that reflection is optimized.
type handlerTypeData struct {
	structValue       reflect.Value
	requestBodyValue  reflect.Value
	responseBodyValue reflect.Value

	isStructPtr   bool
	isRequestPtr  bool
	isResponsePtr bool

	requestFieldNum  int
	responseFieldNum int

	shouldValidateRequest  bool
	shouldValidateResponse bool

	requestMimeType  string
	responseMimeType string
}

func newHandlerTypeData(handler interface{}) handlerTypeData {
	var structValue reflect.Value
	var requestBodyValue reflect.Value
	var responseBodyValue reflect.Value
	var isStructPtr bool
	var isRequestPtr bool
	var isResponsePtr bool
	var requestFieldNum int
	var responseFieldNum int
	var shouldValidateRequest bool
	var shouldValidateResponse bool
	var requestMimeType string
	var responseMimeType string

	structValue = reflect.ValueOf(handler)
	if structValue.Kind() == reflect.Ptr {
		structValue = structValue.Elem()
		isStructPtr = true
	} else {
		panic("handler must be a reference")
	}

	for i := 0; i < structValue.NumField(); i++ {
		structFieldValue := structValue.Field(i)
		structField := structValue.Type().Field(i)
		if tagValue, exists := structField.Tag.Lookup(structTagKey); exists {
			tagValueParts := strings.Split(tagValue, ",")

			// Detect mime type.
			var mimeType string
			for i := 1; i < len(tagValueParts); i++ {
				tagValuePart := tagValueParts[i]

				if strings.HasPrefix(tagValuePart, structTagMimeType+"=") {
					mimeTagValue := strings.SplitN(tagValuePart, "=", 2)
					mimeType = mimeTagValue[1]
					break
				}
			}

			switch tagValueParts[0] {
			case structTagValueRequest:
				requestBodyValue = getFieldValue(structFieldValue)
				requestFieldNum = i
				isRequestPtr = structFieldValue.Kind() == reflect.Ptr
				shouldValidateRequest = hasStructTagOption(tagValue, structTagOptionValidate)
				requestMimeType = mimeType
			case structTagValueResponse:
				responseBodyValue = getFieldValue(structFieldValue)
				responseFieldNum = i
				isResponsePtr = structFieldValue.Kind() == reflect.Ptr
				shouldValidateResponse = hasStructTagOption(tagValue, structTagOptionValidate)
				responseMimeType = mimeType
			}

		}
	}

	return handlerTypeData{
		structValue:            structValue,
		requestBodyValue:       requestBodyValue,
		responseBodyValue:      responseBodyValue,
		isStructPtr:            isStructPtr,
		isRequestPtr:           isRequestPtr,
		isResponsePtr:          isResponsePtr,
		requestFieldNum:        requestFieldNum,
		responseFieldNum:       responseFieldNum,
		shouldValidateRequest:  shouldValidateRequest,
		shouldValidateResponse: shouldValidateResponse,
		requestMimeType:        requestMimeType,
		responseMimeType:       responseMimeType,
	}

}

func (self handlerTypeData) newHandler() interface{} {
	if !self.isStructPtr && !self.isRequestPtr && !self.isResponsePtr {
		handlerCopy := self.structValue.Interface()
		return handlerCopy
	}

	var handlerCopy reflect.Value
	if self.isStructPtr {
		handlerCopy = reflect.New(self.structValue.Type())
	} else {
		handlerCopy = reflect.ValueOf(self.structValue)
	}

	setNewFieldValue(handlerCopy, self.requestFieldNum, self.requestBodyValue)
	setNewFieldValue(handlerCopy, self.responseFieldNum, self.responseBodyValue)

	return handlerCopy.Interface()
}

func setNewFieldValue(structValue reflect.Value, fieldNum int, value reflect.Value) {
	// Only if the value is a pointer. Values will be zero initialized automatically.
	if value.Kind() == reflect.Ptr {
		requestBodyValue := structValue.Elem().Field(fieldNum)
		requestBodyValue.Set(reflect.New(value.Elem().Type()))
	}
}

func getFieldValue(structValue reflect.Value) reflect.Value {
	if structValue.IsZero() && structValue.Kind() == reflect.Ptr {
		newStructFieldValue := reflect.New(structValue.Type().Elem())
		structValue.Set(newStructFieldValue)
	}
	return structValue
}

func hasStructTagOption(structTag string, tagOption string) bool {
	for _, tagPart := range strings.Split(structTag, ",") {
		if tagPart == tagOption {
			return true
		}
	}

	return false
}
