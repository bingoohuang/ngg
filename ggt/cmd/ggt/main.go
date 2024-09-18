package main

import (
	_ "github.com/bingoohuang/ngg/daemon/autoload"
	_ "github.com/bingoohuang/ngg/ggt/ghash"
	_ "github.com/bingoohuang/ngg/ggt/hertz"
	_ "github.com/bingoohuang/ngg/ggt/ip"
	_ "github.com/bingoohuang/ngg/ggt/proxytarget"
	_ "github.com/bingoohuang/ngg/ggt/ps"
	"github.com/bingoohuang/ngg/ggt/root"
	_ "github.com/bingoohuang/ngg/ggt/tecentcloud"
	_ "github.com/bingoohuang/ngg/rotatefile/stdlog/autoload"
)

func main() {
	root.Run()
}
