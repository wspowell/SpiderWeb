package livevalue

import (
	"reflect"
	"sync"
)

// type Listener interface {
// 	// Listen for changes to a value and receives it when it does.
// 	// Implementations must send
// 	Listen() <-chan interface{}
// 	Load(value interface{})
// 	Close()
// }

type UpdateFunc func(valueChan chan<- interface{})

type Value struct {
	mutex        sync.RWMutex
	valueChan    chan interface{}
	currentValue interface{}
}

func NewValue() *Value {
	value := &Value{
		mutex:     sync.RWMutex{},
		valueChan: make(chan interface{}),
	}

	go func() {
		for {
			newValue := <-value.valueChan
			value.mutex.Lock()
			value.currentValue = newValue
			value.mutex.Unlock()
		}
	}()

	return value
}

func (self *Value) Close() error {
	close(self.valueChan)
	return nil
}

func (self *Value) Load(value interface{}) {
	self.mutex.RLock()
	defer self.mutex.RUnlock()

	ptrRef := reflect.ValueOf(value)
	if ptrRef.Kind() != reflect.Ptr {
		panic("value must be pointer")
	}
	ref := ptrRef.Elem()
	ref.Set(reflect.ValueOf(self.currentValue).Elem())
}

func (self *Value) Listen(update UpdateFunc) {
	update(self.valueChan)
}
