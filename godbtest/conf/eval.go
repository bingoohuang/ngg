package conf

import (
	"github.com/Pallinder/go-randomdata"
	"github.com/bingoohuang/ngg/jj"
)

var substituteFns = jj.DefaultSubstituteFns

func init() {
	substituteFns["address"] = func(args string) any {
		// Print an american sounding address
		return randomdata.Address()
	}

	substituteFns["state"] = func(args string) any {
		return randomdata.State(randomdata.Large)
	}
}
