package switchboard

import (
	"reflect"
	"sync"

	"github.com/wspowell/context"
	"github.com/wspowell/log"
)

type ListenFunc func(ctx context.Context, name string, value any)
type UpdateFunc func(ctx context.Context, name string, value Setter)

type Value interface {
	Setter

	Value() any
	Listen(listenFn ListenFunc)
}

type Setter interface {
	Set(ctx context.Context, value any) bool
}

type liveValue struct {
	mutex     *sync.RWMutex
	name      string
	value     any
	listeners []ListenFunc
}

func NewValue(ctx context.Context, name string, defaultValue any, updateFn UpdateFunc) Value {
	value := &liveValue{
		mutex:     &sync.RWMutex{},
		name:      name,
		value:     defaultValue,
		listeners: []ListenFunc{},
	}

	go func(ctx context.Context) {
		ctx = context.Localize(ctx)
		log.Tag(ctx, "switchboard_value_name", name)

		updateFn(ctx, name, value)
	}(ctx)

	return value
}

func (self *liveValue) Set(ctx context.Context, value any) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if !reflect.DeepEqual(self.value, value) {
		log.Debug(ctx, "value changed: %+v -> %+v", self.value, value)
		self.value = value

		for _, listenerFn := range self.listeners {
			go func(ctx context.Context, listenerFn ListenFunc) {
				ctx = context.Localize(ctx)
				listenerFn(ctx, self.name, value)
			}(ctx, listenerFn)
		}

		return true
	}

	return false
}

func (self *liveValue) Value() any {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	return self.value
}

func (self *liveValue) Listen(listenFn ListenFunc) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	self.listeners = append(self.listeners, listenFn)
}
