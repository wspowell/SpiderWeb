package switchboard_test

// import (
// 	"sync"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/wspowell/context"
// 	"github.com/wspowell/log"
// 	"github.com/wspowell/spiderweb/switchboard"
// )

// func Test_LiveValue_Value(t *testing.T) {
// 	ctx := log.WithContext(context.Local(), log.NewConfig(log.LevelDebug))

// 	// Create a new value.
// 	baseValue := switchboard.NewLiveValue("Test_LiveValue_Value")

// 	value := switchboard.Value(baseValue)
// 	assert.Equal(t, "Test_LiveValue_Value", value.Value().(string))

// 	assert.True(t, baseValue.Set(ctx, "new"))

// 	assert.Equal(t, "new", value.Value())
// }

// func Test_LiveValue_Listen(t *testing.T) {
// 	ctx := log.WithContext(context.Local(), log.NewConfig(log.LevelDebug))

// 	// Create a new value.
// 	baseValue := switchboard.NewLiveValue("Test_LiveValue_Value")

// 	value := switchboard.Value(baseValue)
// 	assert.Equal(t, "Test_LiveValue_Value", value.Value().(string))

// 	wg := &sync.WaitGroup{}
// 	wg.Add(1)
// 	value.Listen(func(value any) {
// 		defer wg.Done()
// 		assert.Equal(t, "new", value)
// 	})

// 	assert.True(t, baseValue.Set(ctx, "new"))

// 	wg.Wait()
// }

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wspowell/context"
	"github.com/wspowell/log"

	"github.com/wspowell/spiderweb/switchboard"
)

func noopFn(ctx context.Context, name string, value switchboard.Setter) {}

func updateFn(ctx context.Context, name string, value switchboard.Setter) {
	ticker := time.NewTicker(time.Millisecond)
	for {
		<-ticker.C
		value.Set(ctx, "new")
	}
}

func Test_Value_Set(t *testing.T) {
	t.Parallel()

	ctx := log.WithContext(context.Local(), log.NewConfig(log.LevelDebug))

	// Create a new value.
	value := switchboard.NewValue(ctx, "test", "Test_Value_Set", noopFn)

	assert.Equal(t, "Test_Value_Set", value.Value().(string))
	assert.True(t, value.Set(ctx, "new"))
	assert.Equal(t, "new", value.Value())
}

func Test_Value_Set_no_change(t *testing.T) {
	t.Parallel()

	ctx := log.WithContext(context.Local(), log.NewConfig(log.LevelDebug))

	// Create a new value.
	value := switchboard.NewValue(ctx, "test", "Test_Value_Set", noopFn)

	assert.Equal(t, "Test_Value_Set", value.Value().(string))
	assert.False(t, value.Set(ctx, "Test_Value_Set"))
	assert.Equal(t, "Test_Value_Set", value.Value())
}

func Test_Value_UpdateFunc(t *testing.T) {
	t.Parallel()

	ctx := log.WithContext(context.Local(), log.NewConfig(log.LevelDebug))

	// Create a new value.
	value := switchboard.NewValue(ctx, "test", "Test_Value_UpdateFunc", updateFn)

	time.Sleep(time.Second)
	assert.Equal(t, "new", value.Value())
}

func Test_Value_Listen(t *testing.T) {
	t.Parallel()

	ctx := log.WithContext(context.Local(), log.NewConfig(log.LevelDebug))

	// Create a new value.
	value := switchboard.NewValue(ctx, "test", "Test_Value_Listen", noopFn)

	assert.Equal(t, "Test_Value_Listen", value.Value().(string))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	value.Listen(func(ctx context.Context, name string, value any) {
		defer wg.Done()

		assert.Equal(t, "test", name)
		assert.Equal(t, "new", value)
	})

	assert.True(t, value.Set(ctx, "new"))

	wg.Wait()
}
