package local_test

import (
	"github.com/wspowell/spiderweb/local"
	"testing"
)

func Benchmark_Localized(b *testing.B) {

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			localCtx := local.NewLocalized()
			localCtx.Localize("key", "value")
			localCtx.Value("key")
		}
	})
}
