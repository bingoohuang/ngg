package main

import (
	_ "github.com/bingoohuang/ngg/daemon/autoload"
	"github.com/bingoohuang/ngg/ggt/root"
	_ "github.com/bingoohuang/ngg/rotatefile/stdlog/autoload"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	root.Run()
}
