package ss_test

import (
	"fmt"

	"github.com/bingoohuang/ngg/ss"
)

func ExampleIf() {
	fmt.Println(ss.If(true, "bingoo", "huang"))
	fmt.Println(ss.If(false, "bingoo", "huang"))
	// Output:
	// bingoo
	// huang
}
