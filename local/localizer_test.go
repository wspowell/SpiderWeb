package local_test

import (
	"testing"

	"spiderweb/local"
)

func Test_Localized(t *testing.T) {
	localCtx := local.NewLocalized()
	local.WithValue(localCtx, "ctxKey", "ctxValue")
	local.WithValue(localCtx, "duplicatedKey", "duplicatedValueCtx")
	localCtx.Localize("localKey", "localValue")
	localCtx.Localize("duplicatedKey", "duplicatedValueLocal")

	if localCtx.Value("ctxKey") != "ctxValue" {
		t.Errorf("expected 'ctxKey' to be %v but was %v", "ctxValue", localCtx.Value("ctxKey"))
	}

	if localCtx.Value("localKey") != "localValue" {
		t.Errorf("expected 'localKey' to be %v but was %v", "localValue", localCtx.Value("localKey"))
	}

	if localCtx.Value("duplicatedKey") != "duplicatedValueLocal" {
		t.Errorf("expected 'duplicatedKey' to be %v but was %v", "duplicatedValueLocal", localCtx.Value("duplicatedKey"))
	}
}

// type myLocalizedCtx struct {
// 	local.RequestLocalizer

// 	myHeader string
// }

// func newMyLocalizedCtx(localizer local.RequestLocalizer) *myLocalizedCtx {
// 	var myHeader string
// 	if myHeaders, exists := localizer.Request().Header["X-My-Header"]; exists {
// 		myHeader = myHeaders[0]
// 	}

// 	return &myLocalizedCtx{
// 		RequestLocalizer: localizer,
// 		myHeader:         myHeader,
// 	}
// }

// func Test_Localizer_Wrapped(t *testing.T) {
// 	request := httptest.NewRequest(http.MethodGet, "/myresource", nil)
// 	request.Header["X-My-Header"] = []string{"test value"}

// 	localizedRequest := local.NewLocalizedRequest(request)
// 	localizedRequest.Localize("localKey", "localValue")

// 	myLocalCtx := newMyLocalizedCtx(localizedRequest)

// 	if myLocalCtx.myHeader != "test value" {
// 		t.Errorf("expected 'myHeader' to be %v but was %v", "test value", myLocalCtx.myHeader)
// 	}

// 	if localValue, ok := myLocalCtx.Value("localKey").(string); !ok || localValue != "localValue" {
// 		t.Errorf("expected localized value 'localKey' to be 'localValue'")
// 	}
// }
