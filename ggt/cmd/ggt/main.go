package main

import (
	_ "github.com/bingoohuang/ngg/ggt/ghash"
	_ "github.com/bingoohuang/ngg/ggt/hertz"
	_ "github.com/bingoohuang/ngg/ggt/ip"
	"github.com/bingoohuang/ngg/ggt/root"
	_ "github.com/bingoohuang/ngg/rotatefile/stdlog/autoload"
)

func main() {
	root.Run()
}
