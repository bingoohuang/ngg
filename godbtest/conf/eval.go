package conf

import (
	"github.com/Pallinder/go-randomdata"
	"github.com/bingoohuang/ngg/jj"
)

var substituteFns = jj.DefaultSubstituteFns

func init() {
	substituteFns["address"] = jj.SubstituteFn{
		Fn: func(args string) any {
			// Print an american sounding address
			return randomdata.Address()
		},
		Demo: "随机地址: e.g. @address",
	}

	substituteFns["state"] = jj.SubstituteFn{
		Fn: func(args string) any {
			return randomdata.State(randomdata.Large)
		},
		Demo: "随机州: e.g. @state",
	}
}
