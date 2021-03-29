package logging

import (
	"io"
	"sync"
	"testing"
)

type testConfig struct {
	*Config
}

func (self *testConfig) Out() io.Writer {
	return io.Discard
}

func newTestConfig(level Level, globalTags map[string]interface{}) *testConfig {
	config := NewConfig(level, globalTags)
	return &testConfig{
		Config: config,
	}
}

func Test_Log_GlobalTags_race(t *testing.T) {
	cfg := newTestConfig(LevelError, map[string]interface{}{})
	log := NewLog(cfg)

	wg := &sync.WaitGroup{}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			wg.Done()
			log.Error("test")
		}()
	}

	wg.Wait()
}

func Test_Log_Tags(t *testing.T) {
	cfg := newTestConfig(LevelError, map[string]interface{}{})
	log := NewLog(cfg)

	log.Tag("test1", "value1")
	log.Tag("test2", "value2")

	if value, ok := log.tags["test1"]; !ok || value != "value1" {
		t.Fail()
	}

	if value, ok := log.tags["test2"]; !ok || value != "value2" {
		t.Fail()
	}
}
