package async

import (
	"sync"
	"sync/atomic"
	"time"
)

type Future[Output any] struct {
	outputChannel chan Output
}

func newFuture[Output any]() Future[Output] {
	return Future[Output]{
		outputChannel: make(chan Output),
	}
}

func (self Future[Output]) close() {
	close(self.outputChannel)
}

func (self Future[Output]) Await() Output {
	return <-self.outputChannel
}

type processFunc[Input any, Output any] func(input Input) Output

func Fn[Input any, Output any](process processFunc[Input, Output]) *fnRunner[Input, Output] {
	return newFnRunner(process)
}

type fnRunner[Input any, Output any] struct {
	numGoroutines *int64
	maxGoroutines int64
	inputChannel  chan Input
	fn            processFunc[Input, Output]
}

func newFnRunner[Input any, Output any](fn processFunc[Input, Output]) *fnRunner[Input, Output] {
	return &fnRunner[Input, Output]{
		fn: fn,
	}
}

func (self *fnRunner[Input, Output]) Run(input Input) Future[Output] {
	future := newFuture[Output]()
	go func() {
		defer future.close()
		future.outputChannel <- self.fn(input)
	}()
	return future
}

// -----------------------------------------

type taskProcess[Input any, Output any] func(input Input) Output

func Task[Input any, Output any](task taskProcess[Input, Output]) *taskRunner[Input, Output] {
	return newTaskRunner(task)
}

type taskRunner[Input any, Output any] struct {
	numGoroutines *int64
	maxGoroutines int64
	inputChannel  chan Input
	outputChannel chan Output
	task          taskProcess[Input, Output]
}

func newTaskRunner[Input any, Output any](task taskProcess[Input, Output]) *taskRunner[Input, Output] {
	numGoroutines := int64(0)
	maxGoroutines := int64(10)
	return &taskRunner[Input, Output]{
		numGoroutines: &numGoroutines,
		maxGoroutines: maxGoroutines,
		inputChannel:  make(chan Input),
		outputChannel: make(chan Output),
		task:          task,
	}
}

func (self *taskRunner[Input, Output]) Close() {
	close(self.inputChannel)
	close(self.outputChannel)
}

func (self *taskRunner[Input, Output]) Run(input Input) <-chan Output {
	currentNumGoroutines := atomic.AddInt64(self.numGoroutines, 1)

	if currentNumGoroutines < self.maxGoroutines {
		self.startWorker()
	} else {
		atomic.AddInt64(self.numGoroutines, -1)
	}

	self.inputChannel <- input
	return self.outputChannel
}

func (self *taskRunner[Input, Output]) startWorker() {
	go func() {
		defer atomic.AddInt64(self.numGoroutines, -1)

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		var input Input
		var open bool
		for {
			select {
			case <-ticker.C:
				return
			case input, open = <-self.inputChannel:
				if !open {
					return
				}

				ticker.Stop()
				self.outputChannel <- self.task(input)
				ticker.Reset(time.Second)
			}
		}
	}()
}

// ---------------------------------------------------

func CachedTask[Input comparable, Output any](task taskProcess[Input, Output]) *cachedTaskRunner[Input, Output] {
	return newCachedTaskRunner(task)
}

type cachedTaskRunner[Input comparable, Output any] struct {
	numGoroutines *int64
	maxGoroutines int64
	inputChannel  chan Input
	outputChannel chan Output
	task          taskProcess[Input, Output]
	cachedResults *sync.Map
}

func newCachedTaskRunner[Input comparable, Output any](task taskProcess[Input, Output]) *cachedTaskRunner[Input, Output] {
	numGoroutines := int64(0)
	maxGoroutines := int64(10)
	return &cachedTaskRunner[Input, Output]{
		numGoroutines: &numGoroutines,
		maxGoroutines: maxGoroutines,
		inputChannel:  make(chan Input),
		outputChannel: make(chan Output),
		task:          task,
		cachedResults: &sync.Map{},
	}
}

func (self *cachedTaskRunner[Input, Output]) Close() {
	close(self.inputChannel)
	close(self.outputChannel)
}

func (self *cachedTaskRunner[Input, Output]) Run(input Input) <-chan Output {
	currentNumGoroutines := atomic.AddInt64(self.numGoroutines, 1)

	if currentNumGoroutines < self.maxGoroutines {
		self.startWorker()
	} else {
		atomic.AddInt64(self.numGoroutines, -1)
	}

	self.inputChannel <- input
	return self.outputChannel
}

func (self *cachedTaskRunner[Input, Output]) startWorker() {
	go func() {
		defer atomic.AddInt64(self.numGoroutines, -1)

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		var input Input
		var open bool
		for {
			select {
			case <-ticker.C:
				return
			case input, open = <-self.inputChannel:
				if !open {
					return
				}

				ticker.Stop()
				output, cached := self.cachedResults.Load(input)
				if !cached {
					output = self.task(input)
					self.cachedResults.Store(input, output.(Output))
				}

				self.outputChannel <- output.(Output)
				ticker.Reset(time.Second)
			}
		}
	}()
}

// func Run[Input any, Output any](task taskRunner[Input, Output], input Input) <-chan Output {
// 	value, loaded := taskRunners.Load(task.id)
// 	if !loaded {
// 		newRunner := newTaskRunner(task.task)

// 		value, loaded = taskRunners.LoadOrStore(task.id, newRunner)

// 		if loaded {
// 			// Clean up the unnecessary channels we just created.
// 			newRunner.Close()

// 			// Send the data.
// 			runner := value.(*taskRunner[Input, Output])
// 			newRunner.Run(input)
// 			return runner.outputChannel
// 		}
// 	}

// 	// New task was stored so start the worker.

// 	runner := value.(*taskRunner[Input, Output])

// 	runner.Run(input)
// 	return runner.outputChannel
// }

// type Task[Input any, Output any, Process Processor[Input, Output]] struct {
// 	processor Process
// }

// func NewTask[Input any, Output any, Process Processor[Input, Output]](processor Process) *Task[Input, Output, Process] {
// 	return &Task[Input, Output, Process]{
// 		processor: processor,
// 	}
// }

// func (self *Task[Input, Output, Process]) Process(input Input) Output {
// 	return self.processor(input)
// }

// // Process input.
// 	// ...
// 	fmt.Println("from task", input)

// 	// Send output.
// 	var output Output
// 	return output
