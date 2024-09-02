package root

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type RootCmd struct {
	*cobra.Command
}

func create() *RootCmd {
	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:   "ggt",
		Short: "golang tools",
	}

	r := &RootCmd{Command: rootCmd}
	return r
}

var Cmd = create()

func Run() {
	if err := Cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}
