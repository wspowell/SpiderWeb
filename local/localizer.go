package local

import (
	"context"
	"time"
)

// contextualizer is a context.Context that allows the underlying immutable
//   context.Context to be accessed and overridden.
type contextualizer interface {
	// Context embedded behavior.
	context.Context

	// SetContext sets the underlying context.Context.
	setContext(context.Context)

	// GetContext returns the underlying context.Context.
	// Returned value must be thread safe.
	Context() context.Context
}

var _ contextualizer = (*contextualized)(nil)

type contextualized struct {
	context context.Context
}

func (self *contextualized) Deadline() (deadline time.Time, ok bool) {
	return self.context.Deadline()
}

func (self *contextualized) Done() <-chan struct{} {
	return self.context.Done()
}

func (self *contextualized) Err() error {
	return self.context.Err()
}

func (self *contextualized) Value(key interface{}) interface{} {
	return self.context.Value(key)
}

func (self *contextualized) Context() context.Context {
	return self.context
}

func (self *contextualized) setContext(context context.Context) {
	self.context = context
}

// Context is a alias for Localizer.
// Using local.Context may read better and help associated it with context.Context.
type Context Localizer

// Localizer stores data local to a goroutine. A localized context.
// This works differently than context.Context in that it is not meant to
//   cross API boundaries and is not immutable. However, it is designed to
//   be able to work alongside context.Context. It is also meant to be
//   wrapped by developers to allow for direct access of endpoint local data.
// Localizer is also a context.Context. It is expected that the value of
//   the internal context.Context is thread safe.
// Not thread safe.
type Localizer interface {
	contextualizer
	Localize(key interface{}, value interface{})
}

var _ Localizer = (*Localized)(nil)

type Localized struct {
	contextualized

	// Store locals in a map that do not have a defined variable.
	locals map[interface{}]interface{}
}

func NewLocalized() *Localized {
	return &Localized{
		contextualized: contextualized{context.Background()},
		locals:         map[interface{}]interface{}{},
	}
}

func WithContext(context context.Context) *Localized {
	return &Localized{
		contextualized: contextualized{context},
		locals:         map[interface{}]interface{}{},
	}
}

func (self *Localized) Value(key interface{}) interface{} {
	if localValue, exists := self.locals[key]; exists {
		return localValue
	}

	return self.context.Value(key)
}

func (self *Localized) Localize(key interface{}, value interface{}) {
	self.locals[key] = value
}

func WithValue(parent Localizer, key interface{}, value interface{}) {
	childContext := context.WithValue(parent.Context(), key, value)

	parent.setContext(childContext)
}

func WithCancel(parent Localizer) context.CancelFunc {
	childContext, cancel := context.WithCancel(parent.Context())

	parent.setContext(childContext)

	return cancel
}

func WithDeadline(parent Localizer, deadline time.Time) context.CancelFunc {
	childContext, cancel := context.WithDeadline(parent.Context(), deadline)

	parent.setContext(childContext)

	return cancel
}

func WithTimeout(parent Localizer, timeout time.Duration) context.CancelFunc {
	childContext, cancel := context.WithTimeout(parent.Context(), timeout)

	parent.setContext(childContext)

	return cancel
}
