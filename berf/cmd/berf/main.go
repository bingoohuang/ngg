package main

import (
	"github.com/bingoohuang/ngg/berf/pkg/blow"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/pflag"
)

func init() {
	pflag.Parse()
}

func main() {
	blow.StartBlow()
}
