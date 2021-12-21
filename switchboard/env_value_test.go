package switchboard_test

// import (
// 	"os"
// 	"sync"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/wspowell/context"
// 	"github.com/wspowell/log"
// 	"github.com/wspowell/spiderweb/switchboard"
// )

// func Test_EnvValue_Value(t *testing.T) {
// 	ctx := log.WithContext(context.Background(), log.NewConfig().WithLevel(log.LevelDebug))

// 	// Create a new value.
// 	envValue := switchboard.NewEnvValue(ctx, "Test_EnvValue_Value", "default", time.Millisecond)

// 	value := switchboard.Value(envValue)
// 	assert.Equal(t, "default", value.Value().(string))

// 	os.Setenv("Test_EnvValue_Value", "new")
// 	time.Sleep(2 * time.Millisecond)

// 	assert.Equal(t, "new", value.Value().(string))
// }

// func Test_EnvValue_Wait(t *testing.T) {
// 	ctx := log.WithContext(context.Background(), log.NewConfig().WithLevel(log.LevelDebug))

// 	// Create a new value.
// 	envValue := switchboard.NewEnvValue(ctx, "Test_EnvValue_Wait", "default", time.Millisecond)

// 	value := switchboard.Value(envValue)
// 	assert.Equal(t, "default", value.Value().(string))

// 	os.Setenv("Test_EnvValue_Wait", "new")

// 	wg := &sync.WaitGroup{}
// 	wg.Add(1)

// 	value.Listen(func(value any) {
// 		defer wg.Done()
// 		assert.Equal(t, "new", value.(string))
// 	})

// 	wg.Wait()

// 	assert.Equal(t, "new", value.Value().(string))
// }

// func Test_EnvValue_context_cancel(t *testing.T) {
// 	ctx := log.WithContext(context.Background(), log.NewConfig().WithLevel(log.LevelDebug))

// 	ctx, cancel := context.WithCancel(ctx)

// 	// Create a new value.
// 	switchboard.NewEnvValue(ctx, "Test_EnvValue_context_cancel", "default", time.Millisecond)

// 	// Confirmed working by coverage.
// 	cancel()

// 	time.Sleep(time.Second)
// }

// func Test_EnvValue_config_example(t *testing.T) {
// 	type config struct {
// 		foo *switchboard.EnvValue
// 		bar *switchboard.EnvValue
// 		zar *switchboard.EnvValue
// 	}

// 	ctx := log.WithContext(context.Background(), log.NewConfig().WithLevel(log.LevelDebug))

// 	cfg := &config{
// 		foo: switchboard.NewEnvValue(ctx, "foo", "foo_default", time.Millisecond),
// 		bar: switchboard.NewEnvValue(ctx, "bar", "bar_default", time.Millisecond),
// 		zar: switchboard.NewEnvValue(ctx, "zar", "zar_default", time.Millisecond),
// 	}

// 	wg := &sync.WaitGroup{}
// 	wg.Add(3)

// 	cfg.foo.Listen(func(value any) {
// 		defer wg.Done()
// 		assert.Equal(t, "foo_new", value)
// 	})
// 	cfg.bar.Listen(func(value any) {
// 		defer wg.Done()
// 		assert.Equal(t, "bar_new", value)
// 	})
// 	cfg.zar.Listen(func(value any) {
// 		defer wg.Done()
// 		assert.Equal(t, "zar_new", value)
// 	})

// 	go func() {
// 		os.Setenv("foo", "foo_new")
// 		os.Setenv("bar", "bar_new")
// 		os.Setenv("zar", "zar_new")
// 	}()

// 	wg.Wait()
// }
