package tick

import (
	"os"
	"time"

	"github.com/bingoohuang/ngg/ss"
)

func Getenv(str string, defaultValue time.Duration) (time.Duration, error) {
	env := os.Getenv(str)
	if env == "" {
		return defaultValue, nil
	}

	val, _, err := ss.ParseDur(env)
	if err != nil {
		return 0, err
	}
	return val, nil
}
