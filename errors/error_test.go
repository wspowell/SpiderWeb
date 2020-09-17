package errors

import (
	"fmt"
	"testing"
)

func Test_New(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		about        string
		internalCode string
		err          string
	}{
		{"it creates a new internal error with the given error", "ER1001", "whoops"},
		{"it creates a new internal error when internal code is empty", "", "whoops"},
		{"it creates a new internal error when err is nil", "ER1001", ""},
	}
	for _, testCase := range testCases {
		t.Run(testCase.about, func(t *testing.T) {
			err := New(testCase.internalCode, testCase.err)

			if err == nil {
				t.Errorf("created internal error is nil")
				return
			}

			var expectedFeatureCode string

			if testCase.internalCode == "" {
				expectedFeatureCode = ""
			} else {
				expectedFeatureCode = testCase.internalCode
			}

			expectedErrorString := testCase.err

			if err.(internalError).internalCodeStack[0] != expectedFeatureCode {
				t.Errorf("expected internal code '%v', got '%v'", testCase.err, err)
				return
			}

			if err.Error() != expectedErrorString {
				t.Errorf("expected error '%v', got '%v'", testCase.err, err)
				return
			}
		})
	}
}

func Test_Wrap(t *testing.T) {
	testCases := []struct {
		about        string
		internalCode string
		err          error
	}{
		{"it creates a new internal error with the given error", "ER1001", fmt.Errorf("whoops")},
		{"it creates a new internal error when internal code is empty", "", fmt.Errorf("whoops")},
		//{"it creates a new internal error when err is nil", "ER1001", nil},
	}
	for _, testCase := range testCases {
		t.Run(testCase.about, func(t *testing.T) {
			err := Wrap(testCase.internalCode, testCase.err)

			if err == nil {
				t.Errorf("created internal error is nil")
				return
			}

			var expectedFeatureCode string
			var expectedErrorString string

			if testCase.internalCode == "" {
				expectedFeatureCode = ""
			} else {
				expectedFeatureCode = testCase.internalCode
			}

			if testCase.err == nil {
				expectedErrorString = emptyErrorMessage
			} else {
				expectedErrorString = testCase.err.Error()
			}

			if err.(internalError).internalCodeStack[0] != expectedFeatureCode {
				t.Errorf("expected internal code '%v', got '%v'", testCase.err.Error(), err.Error())
				return
			}

			if err.Error() != expectedErrorString {
				t.Errorf("expected error '%v', got '%v'", expectedErrorString, err.Error())
				return
			}
		})
	}
}
