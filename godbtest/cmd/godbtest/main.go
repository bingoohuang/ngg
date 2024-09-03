package main

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/bingoohuang/ngg/daemon/autoload"
	"github.com/bingoohuang/ngg/godbtest"
	"github.com/bingoohuang/ngg/godbtest/conf"
	_ "github.com/bingoohuang/ngg/godbtest/drivers"
	"github.com/bingoohuang/ngg/ver"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/pflag"
)

func main() {
	c, err := conf.ParseConfigFile(configFile, driverName)
	if err != nil {
		log.Fatal(err)
	}
	c.Go(dsn, evaluates, pflag.Args())
}

var (
	configFile string
	driverName string
	dsn        string
	evaluates  []string
)

func init() {
	pflag.StringVarP(&configFile, "conf", "c", "", "Config file path, or new to create demo one")
	pflag.StringVarP(&driverName, "driver", "d", "", "Part of driver name to filter, e.g. ora, my")
	pflag.StringVarP(&dsn, "dsn", "", os.Getenv("DSN"), "Data source name")
	pflag.StringArrayVarP(&evaluates, "evaluate", "e", nil, "Evaluate query or query file")
	pVersion := pflag.BoolP("version", "v", false, "show version and exit")
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		pflag.PrintDefaults()

		fmt.Fprintf(os.Stderr, `Examples:
  1. More help: %s -e %%help

`, os.Args[0])
	}
	pflag.Parse()

	if *pVersion {
		fmt.Println(ver.Version())
		os.Exit(0)
	}

	if configFile == "new" {
		configFile = "db-" + time.Now().Format(`20060102150405.yml`)
		if err := os.WriteFile(configFile, godbtest.DemoConf, os.ModePerm); err != nil {
			log.Fatalf("creating demo configuration file: %v", err)
		}
		fmt.Printf("demo config file %s generated\n", configFile)
		os.Exit(0)
	}
}
