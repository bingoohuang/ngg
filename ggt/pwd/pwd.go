package pwd

import (
	"log"
	"os"

	"github.com/atotto/clipboard"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/spf13/cobra"
)

var cobraCmd = &cobra.Command{
	Use:  "pwd",
	Long: "copy current dir to clipboard",
}

func init() {
	root.AddCommand(cobraCmd, &subCmd{})
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
