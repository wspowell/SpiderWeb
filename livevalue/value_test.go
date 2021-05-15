package livevalue_test

import (
	"os"
	"testing"
	"time"

	"github.com/wspowell/spiderweb/livevalue"

	"github.com/stretchr/testify/assert"
)

func Test_Value_EnvListener(t *testing.T) {
	// Simulate pre-existing environment.
	os.Setenv("ttl", "10s")
	os.Setenv("max_value", "13")

	type envConfig struct {
		Ttl      time.Duration `env:"ttl" envDefault:"5s"`
		Flag     bool          `env:"flag"`
		MaxValue int           `env:"max_value"`
	}

	// Simulate setting up the value in an app.
	cfg := &envConfig{}
	envValue := livevalue.NewEnvValue(cfg, 1*time.Second)

	// Check default value.
	assert.Equal(t, &envConfig{
		Ttl:      10 * time.Second,
		Flag:     false,
		MaxValue: 13,
	}, cfg)

	// Simulate a process accessing the value.
	cfg = &envConfig{}
	envValue.Load(cfg)

	assert.Equal(t, &envConfig{
		Ttl:      10 * time.Second,
		Flag:     false,
		MaxValue: 13,
	}, cfg)

	// Update the environment and wait for an update.
	os.Setenv("flag", "true")
	os.Setenv("max_value", "9")

	time.Sleep(2 * time.Second)

	cfg = &envConfig{}
	envValue.Load(cfg)

	assert.Equal(t, &envConfig{
		Ttl:      10 * time.Second,
		Flag:     true,
		MaxValue: 9,
	}, cfg)

}
