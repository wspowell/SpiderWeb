package logging

import "testing"

func Benchmark_debug_at_error_level(b *testing.B) {
	log := NewLog(newTestConfig(LevelError, map[string]interface{}{}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Debug("hello: %s", "world")
	}
}

func Benchmark_error_at_error_level(b *testing.B) {
	log := NewLog(newTestConfig(LevelError, map[string]interface{}{}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Error("hello: %s", "world")
	}
}

func Benchmark_debug_at_error_level_1000x(b *testing.B) {
	log := NewLog(newTestConfig(LevelError, map[string]interface{}{}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for k := 0; k < 1000; k++ {
			log.Debug("hello: %s", "world")
		}
	}
}

func Benchmark_error_at_error_level_1000x(b *testing.B) {
	log := NewLog(newTestConfig(LevelError, map[string]interface{}{}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for k := 0; k < 1000; k++ {
			log.Error("hello: %s", "world")
		}
	}
}

func Benchmark_NewLog(b *testing.B) {
	config := NewConfig(LevelError, map[string]interface{}{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewLog(config)
	}
}

func Benchmark_ConfigCopy_NewLog(b *testing.B) {
	config := NewConfig(LevelError, map[string]interface{}{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := config
		NewLog(cfg)
	}
}
