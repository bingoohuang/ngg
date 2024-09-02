package main

import (
	"fmt"
	"github.com/bingoohuang/ngg/berf/pkg/blow"
	"github.com/bingoohuang/ngg/ver"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/pflag"
	"os"
)

func init() {
	pVersion := pflag.Bool("version", false, "show version info")
	pflag.Parse()
	if *pVersion {
		fmt.Println(ver.Version())
		os.Exit(0)
	}
}

func main() {
	blow.StartBlow()
}
