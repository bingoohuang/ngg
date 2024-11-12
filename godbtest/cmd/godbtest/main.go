package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/godbtest"
	"github.com/bingoohuang/ngg/godbtest/conf"
	_ "github.com/bingoohuang/ngg/godbtest/drivers"
	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/ver"
	"github.com/spf13/cobra"
)

var cmd = func() *cobra.Command {
	r := &cobra.Command{
		Use:  "godbtest",
		Long: "use ggt db -e %help for more usages",
	}

	r.Version = "version"
	r.SetVersionTemplate(ver.Version() + "\n")
	return r
}()

type subCmd struct {
	Conf      string   `short:"c" help:"config file path, or new to create demo one"`
	Driver    string   `short:"d" help:"driver name to filter, e.g. ora, my"`
	Dsn       string   `help:"data source name" env:"auto"`
	Evaluates []string `short:"e" help:"Evaluate query or file"`
}

func (fc *subCmd) Run(_ *cobra.Command, args []string) error {
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

func main() {
	RootCommand(cmd, &subCmd{})

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

func RootCommand(c *cobra.Command, fc any) {
	if fc != nil && !c.DisableFlagParsing {
		ss.PanicErr(root.InitFlags(fc, c.Flags(), c.PersistentFlags()))
	}
	if r, ok := fc.(interface {
		Run(cmd *cobra.Command, args []string) error
	}); ok {
		c.Run = func(cmd *cobra.Command, args []string) {
			if err := r.Run(cmd, args); err != nil {
				log.Printf("error occured: %v", err)
			}
		}
	}
}
