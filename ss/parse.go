package ss

import (
	"fmt"
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
