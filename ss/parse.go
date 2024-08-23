package ss

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Parseable interface {
	~bool | ~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~complex64 | ~complex128
}

func Parse[T Parseable](str string) (T, error) {
	var result T
	_, err := fmt.Sscanf(str, "%v", &result)
	return result, err
}

// ParseBool returns the boolean value represented by the string.
// It accepts 1, t, true, y, yes, on as true with camel case incentive
// and accepts 0, f false, n, no, off as false with camel case incentive
// Any other value returns an error.
func ParseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "1", "t", "true", "y", "yes", "on":
		return true, nil
	case "0", "f", "false", "n", "no", "off":
		return false, nil
	}

	return false, fmt.Errorf("unknown bool %s: %w", s, strconv.ErrSyntax)
}

// Getenv 获取环境变量的值
func Getenv[T Parseable](name string, defaultValue T) (T, error) {
	env := os.Getenv(name)
	if env == "" {
		return defaultValue, nil
	}

	val, err := Parse[T](env)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("parse env %s error: %w", name, err)
	}
	return val, nil
}

func GetenvBool(envName string, defaultValue bool) (bool, error) {
	val := os.Getenv(envName)
	if val == "" {
		return defaultValue, nil
	}

	return ParseBool(val)
}

// GetenvBytes 获得环境变量 name 的值所表示的大小，例如. 30MiB
func GetenvBytes(name string, defaultValue uint64) (uint64, error) {
	env := os.Getenv(name)
	if env == "" {
		return defaultValue, nil
	}

	return ParseBytes(env)
}
