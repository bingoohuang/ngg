package dur_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/bingoohuang/ngg/dur"
	"github.com/stretchr/testify/assert"
)

func ExampleRound() {
	for i := 0; i < 15; i++ {
		d := time.Duration(3.455555 * math.Pow(10, float64(i)))
		fmt.Printf("%2d  %12v  %6s\n", i, d, dur.Round(d))

		//         original formatted

		// Output:
		//  0           3ns     3ns
		//  1          34ns    34ns
		//  2         345ns   345ns
		//  3       3.455µs  3.46µs
		//  4      34.555µs  34.6µs
		//  5     345.555µs   346µs
		//  6    3.455555ms  3.46ms
		//  7    34.55555ms  34.6ms
		//  8    345.5555ms   346ms
		//  9     3.455555s   3.46s
		// 10     34.55555s   34.6s
		// 11    5m45.5555s   5m46s
		// 12    57m35.555s  57m36s
		// 13   9h35m55.55s  9h35m56s
		// 14   95h59m15.5s  95h59m16s
	}
}

func TestRound(t *testing.T) {
	assert.Equal(t, "3s", dur.Round(3*time.Second).String())
	assert.Equal(t, "3.1s", dur.Round(3*time.Second+100*time.Millisecond).String())
	assert.Equal(t, "3.12s", dur.Round(3*time.Second+123*time.Millisecond).String())
}
