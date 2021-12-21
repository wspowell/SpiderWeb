package switchboard

// import (
// 	"fmt"
// 	"io"
// 	"os"
// 	"os/exec"
// 	"strings"
// 	"time"

// 	"github.com/wspowell/context"
// 	"github.com/wspowell/log"
// )

// type EnvValue struct {
// 	*LiveValue
// }

// func NewEnvValue(ctx context.Context, name string, defaultValue any, updateInterval time.Duration) *EnvValue {
// 	os.Setenv(name, fmt.Sprintf("%v", defaultValue))

// 	envValue := &EnvValue{
// 		LiveValue: NewLiveValue(defaultValue),
// 	}

// 	go func(ctx context.Context) {
// 		ctx = context.Localize(ctx)
// 		log.Tag(ctx, "env_name", name)

// 		ticker := time.NewTicker(updateInterval)
// 		defer ticker.Stop()

// 		for {
// 			select {
// 			case <-ctx.Done():
// 				return
// 			case <-ticker.C:
// 				cmd := exec.Command("bash", "-c", "env")

// 				output, err := cmd.StdoutPipe()
// 				if err != nil {
// 					log.Error(ctx, "err: %v", err)
// 				}

// 				err = cmd.Start()
// 				if err != nil {
// 					log.Error(ctx, "err: %v", err)
// 				}

// 				outputBytes, err := io.ReadAll(output)
// 				if err != nil {
// 					log.Error(ctx, "err: %v", err)
// 				}

// 				err = cmd.Wait()
// 				if err != nil {
// 					log.Error(ctx, "err: %v", err)
// 				}

// 				//log.Debug(ctx, "output: %s", outputBytes)

// 				envvars := strings.Split(string(outputBytes), "\n")
// 				for _, envvar := range envvars {
// 					split := strings.SplitN(envvar, "=", 2)
// 					//log.Debug(ctx, "%v %v", split[0], split[1])
// 					if split[0] == name {
// 						envValue.Set(ctx, split[1])
// 						break
// 					}
// 				}

// 				//envValue.Set(ctx, os.Getenv(name))
// 			}
// 		}
// 	}(ctx)

// 	return envValue
// }
