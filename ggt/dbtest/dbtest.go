package dbtest

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/spf13/cobra"

	"github.com/bingoohuang/ngg/godbtest"
	"github.com/bingoohuang/ngg/godbtest/conf"
	_ "github.com/bingoohuang/ngg/godbtest/drivers"
)

type subCmd struct {
	Conf      string   `short:"c" help:"config file path, or new to create demo one"`
	Driver    string   `short:"d" help:"driver name to filter, e.g. ora, my"`
	Dsn       string   `help:"data source name"`
	Evaluates []string `short:"e" help:"Evaluate query or file"`
}

func init() {
	fc := &subCmd{}
	c := &cobra.Command{
		Use:  "dbtest",
		Long: "use ggt db -e %help for more usages",
		RunE: fc.run,
	}

	root.AddCommand(c, fc)
}

func (fc *subCmd) run(_ *cobra.Command, args []string) error {
	if fc.Conf == "new" {
		fc.Conf = "db-" + time.Now().Format(`20060102150405.yml`)
		if err := os.WriteFile(fc.Conf, godbtest.DemoConf, os.ModePerm); err != nil {
			log.Fatalf("creating demo configuration file: %v", err)
		}
		fmt.Printf("demo config file %s generated\n", fc.Conf)
		os.Exit(0)
	}

	c, err := conf.ParseConfigFile(fc.Conf, fc.Driver)
	if err != nil {
		return err
	}
	c.Go(fc.Dsn, fc.Evaluates, args)
	return nil
}
