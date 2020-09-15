package local_test

import (
	"context"
	"testing"

	"github.com/wspowell/spiderweb/local"
)

func checkContext(t *testing.T, ctx context.Context) {
	if ctx.Value("localKey") != nil {
		t.Errorf("expected 'localKey' to be nil")
	}

	if ctx.Value("duplicatedKey").(string) != "immutableValue" {
		t.Errorf("expected 'duplicatedKey' to be %v but was %v", "immutableValue", ctx.Value("duplicatedKey"))
	}

	if ctx.Value("immutable").(string) != "immutableValue" {
		t.Errorf("expected 'immutable' to be %v but was %v", "immutable", ctx.Value("immutable"))
	}
}

func checkLocal(t *testing.T, ctx local.Context) {
	if ctx.Value("localKey").(string) != "localValue" {
		t.Errorf("expected 'localKey' to be %v but was %v", "localValue", ctx.Value("localKey"))
	}

	if ctx.Value("duplicatedKey").(string) != "myValue" {
		t.Errorf("expected 'duplicatedKey' to be %v but was %v", "myValue", ctx.Value("duplicatedKey"))
	}

	if ctx.Value("immutable").(string) != "immutableValue" {
		t.Errorf("expected 'immutable' to be %v but was %v", "immutable", ctx.Value("immutable"))
	}
}

func Test_NewLocalized(t *testing.T) {
	localCtx := local.NewLocalized()

	local.WithValue(localCtx, "immutable", "immutableValue")
	local.WithValue(localCtx, "duplicatedKey", "immutableValue")

	localCtx.Localize("localKey", "localValue")
	localCtx.Localize("duplicatedKey", "myValue")

	checkContext(t, localCtx.Context())
	checkLocal(t, localCtx)
}

func Test_FromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "immutable", "immutableValue")
	ctx = context.WithValue(ctx, "duplicatedKey", "immutableValue")

	localCtx := local.FromContext(ctx)
	localCtx.Localize("localKey", "localValue")
	localCtx.Localize("duplicatedKey", "myValue")

	checkContext(t, localCtx.Context())
	checkLocal(t, localCtx)
}
