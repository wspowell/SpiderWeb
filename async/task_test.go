package async_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/wspowell/spiderweb/async"
)

func Test_Future(t *testing.T) {

	asyncEcho := async.Fn(echo)

	future := asyncEcho.Run("hello")

	fmt.Println(future.Await())
}

func Benchmark_Fn(b *testing.B) {
	b.SetParallelism(16)

	asyncEcho := async.Fn(echo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if "hello" != asyncEcho.Run("hello").Await() {
			panic("invalid response")
		}
	}

}

// ---------------------------------------------------

func echo(input string) string {
	return input
}

type model struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type marshalInput struct {
	m model
}

type marshalOutput struct {
	bytes []byte
	err   error
}

func marshalJson(input marshalInput) marshalOutput {
	time.Sleep(1 * time.Second)
	bytes, err := json.Marshal(&input.m)
	return marshalOutput{
		bytes: bytes,
		err:   err,
	}
}

type unmarshalInput struct {
	bytes string
	m     model
}

func (a unmarshalInput) Compare(b unmarshalInput) int {

	return 0
}

type unmarshalOutput struct {
	err error
}

func unmarshalJson(input unmarshalInput) unmarshalOutput {
	time.Sleep(1 * time.Second)
	err := json.Unmarshal([]byte(input.bytes), &input.m)
	return unmarshalOutput{
		err: err,
	}
}

func Test_Run_complex(t *testing.T) {
	unmarshal := async.Task(unmarshalJson)
	marshal := async.Task(marshalJson)

	originalBytes := []byte(`{"id":11,"name":"me"}`)
	unmarshalResult := <-unmarshal.Run(unmarshalInput{
		bytes: string(originalBytes),
		m:     model{},
	})
	if unmarshalResult.err != nil {
		panic(unmarshalResult.err)
	}

	marshalResult := <-marshal.Run(marshalInput{
		m: model{},
	})
	if marshalResult.err != nil {
		panic(marshalResult.err)
	}

	if !bytes.Equal(originalBytes, marshalResult.bytes) {
		panic(fmt.Sprintf("bytes not equal '%s' != '%s'\n", originalBytes, marshalResult.bytes))
	}
}

func Test_Run_complex_cached(t *testing.T) {
	unmarshal := async.CachedTask(unmarshalJson)
	marshal := async.CachedTask(marshalJson)

	originalBytes := []byte(`{"id":11,"name":"me"}`)
	unmarshalResult := <-unmarshal.Run(unmarshalInput{
		bytes: string(originalBytes),
		m:     model{},
	})
	if unmarshalResult.err != nil {
		panic(unmarshalResult.err)
	}

	marshalResult := <-marshal.Run(marshalInput{
		m: model{},
	})
	if marshalResult.err != nil {
		panic(marshalResult.err)
	}

	if !bytes.Equal(originalBytes, marshalResult.bytes) {
		panic(fmt.Sprintf("bytes not equal '%s' != '%s'\n", originalBytes, marshalResult.bytes))
	}

	// Do it again

	unmarshalResult = <-unmarshal.Run(unmarshalInput{
		bytes: string(originalBytes),
		m:     model{},
	})
	if unmarshalResult.err != nil {
		panic(unmarshalResult.err)
	}

	marshalResult = <-marshal.Run(marshalInput{
		m: model{},
	})
	if marshalResult.err != nil {
		panic(marshalResult.err)
	}

	if !bytes.Equal(originalBytes, marshalResult.bytes) {
		panic(fmt.Sprintf("bytes not equal '%s' != '%s'\n", originalBytes, marshalResult.bytes))
	}
}

func Test_Run(t *testing.T) {
	echoTask := async.Task(echo)

	outputChannel := echoTask.Run("hello")

	fmt.Println("received output", <-outputChannel)

	t.Fail()
}

func Benchmark_Run(b *testing.B) {
	b.SetParallelism(16)

	echoTask := async.Task(echo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if "hello" != <-echoTask.Run("hello") {
			panic("invalid response")
		}
	}

}

func Benchmark_Run_parallel(b *testing.B) {
	b.SetParallelism(16)

	echoTask := async.Task(echo)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if "hello" != <-echoTask.Run("hello") {
				panic("invalid response")
			}
		}
	})
}

func Benchmark_Run_complex(b *testing.B) {
	b.SetParallelism(16)

	unmarshal := async.Task(unmarshalJson)
	marshal := async.Task(marshalJson)

	originalBytes := []byte(`{"id":11,"name":"me"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		<-unmarshal.Run(unmarshalInput{
			bytes: string(originalBytes),
			m:     model{},
		})

		<-marshal.Run(marshalInput{
			m: model{},
		})
	}

}

func Benchmark_Run_parallel_complex(b *testing.B) {
	b.SetParallelism(16)

	unmarshal := async.Task(unmarshalJson)
	marshal := async.Task(marshalJson)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		originalBytes := []byte(`{"id":11,"name":"me"}`)
		for pb.Next() {
			<-unmarshal.Run(unmarshalInput{
				bytes: string(originalBytes),
				m:     model{},
			})

			<-marshal.Run(marshalInput{
				m: model{},
			})
		}
	})
}

func Benchmark_Run_parallel_complex_cached(b *testing.B) {
	b.SetParallelism(16)

	unmarshal := async.CachedTask(unmarshalJson)
	marshal := async.CachedTask(marshalJson)

	{
		originalBytes := []byte(`{"id":11,"name":"me"}`)
		<-unmarshal.Run(unmarshalInput{
			bytes: string(originalBytes),
			m:     model{},
		})

		<-marshal.Run(marshalInput{
			m: model{},
		})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		originalBytes := []byte(`{"id":11,"name":"me"}`)
		m := model{}
		for pb.Next() {
			<-unmarshal.Run(unmarshalInput{
				bytes: string(originalBytes),
				m:     model{},
			})

			<-marshal.Run(marshalInput{
				m: m,
			})
		}
	})
}

type EchoProcess struct {
	output string
}

func (self *EchoProcess) Process(input string) string {
	self.output = input

	return input
}

// func Test_Task(t *testing.T) {

// 	echo := EchoProcess{
// 		output: "different",
// 	}

// 	job := async.Task[string, string](echo.Process)

// 	output := job("hello")
// 	fmt.Println("received output", output)
// 	fmt.Println(echo)

// 	t.Fail()
// }
