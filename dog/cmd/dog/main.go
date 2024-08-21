package main

import (
	"flag"

	_ "github.com/bingoohuang/ngg/dog/autoload"
)

func main() {
	flag.Parse()
	cgoDemo()

	select {}
}
