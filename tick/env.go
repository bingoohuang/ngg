package tick

import (
	"os"
	"time"
)

func Getenv(str string, defaultValue time.Duration) (time.Duration, error) {
	env := os.Getenv(str)
	if env == "" {
		return defaultValue, nil
	}

	val, _, err := Parse(env)
	if err != nil {
		return 0, err
	}
	return val, nil
}
