package main

import (
	_ "github.com/bingoohuang/ngg/daemon/autoload"
	_ "github.com/bingoohuang/ngg/ggt/dbtest"
	_ "github.com/bingoohuang/ngg/ggt/frp"
	_ "github.com/bingoohuang/ngg/ggt/ghash"
	_ "github.com/bingoohuang/ngg/ggt/gossh"
	_ "github.com/bingoohuang/ngg/ggt/goup"
	_ "github.com/bingoohuang/ngg/ggt/gurl"
	_ "github.com/bingoohuang/ngg/ggt/hertz"
	_ "github.com/bingoohuang/ngg/ggt/ip"
	_ "github.com/bingoohuang/ngg/ggt/ps"
	"github.com/bingoohuang/ngg/ggt/root"
	_ "github.com/bingoohuang/ngg/ggt/rsa"
	_ "github.com/bingoohuang/ngg/ggt/tencentcloud"
	_ "github.com/bingoohuang/ngg/rotatefile/stdlog/autoload"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	root.Run()
}
