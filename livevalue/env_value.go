package livevalue

import (
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
)

type EnvValue struct {
	*Value
	updateFrequency time.Duration
}

func NewEnvValue(envConfig interface{}, updateFrequency time.Duration) *EnvValue {
	value := &EnvValue{
		Value:           NewValue(),
		updateFrequency: updateFrequency,
	}

	value.currentValue = envConfig
	value.Listen(value.UpdateFn)

	return value
}

func (self *EnvValue) UpdateFn(valueChan chan<- interface{}) {
	// Send default value.
	currentEnv := toMap(os.Environ())
	_ = env.Parse(self.currentValue)
	valueChan <- self.currentValue

	go func() {
		defer close(valueChan)

		ticker := time.NewTicker(self.updateFrequency)
		defer ticker.Stop()

		for {
			<-ticker.C

			var shouldUpdate bool

			newEnv := toMap(os.Environ())
			if len(newEnv) != len(currentEnv) {
				shouldUpdate = true
			} else {
				for key, value := range newEnv {
					if currentEnv[key] != value {
						shouldUpdate = true
						break
					}
				}
			}

			if shouldUpdate {
				currentEnv = newEnv
				_ = env.Parse(self.currentValue)
				valueChan <- self.currentValue
			}
		}
	}()
}

// Copied from "github.com/caarlos0/env/v6"
func toMap(env []string) map[string]string {
	r := map[string]string{}
	for _, e := range env {
		p := strings.SplitN(e, "=", 2)
		r[p[0]] = p[1]
	}
	return r
}
