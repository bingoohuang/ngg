package main

import (
	"fmt"
	"log"
	"os"

	"github.com/atotto/clipboard"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/spf13/cobra"
)

func main() {
	c := root.CreateCmd(nil, "ggtpwd", "copy current dir to clipboard", &subCmd{})
	if err := c.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

type subCmd struct {
}

func (f *subCmd) Run(cmd *cobra.Command, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := clipboard.WriteAll(wd); err != nil {
		return err
	}
	log.Printf("%q copied to clipboard", wd)

	return nil
}
