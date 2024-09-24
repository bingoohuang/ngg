package urlx

import (
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/ss"
	"github.com/spf13/cobra"
)

func init() {
	fc := &subCmd{}
	c := &cobra.Command{
		Use:   "size",
		Short: "convert size among bytes and human-readable format",
		RunE:  fc.run,
	}

	root.AddCommand(c, fc)
}

type subCmd struct {
}

func (f *subCmd) run(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		fmt.Println("usage: ggt size 123445 12M")
	}

	for _, arg := range args {
		if regexp.MustCompile(`^\d+$`).MatchString(arg) {
			v, _ := strconv.ParseUint(arg, 10, 64)
			log.Printf("%s => IBytes: %s, Bytes: %s", arg, ss.IBytes(v), ss.Bytes(v))
		} else {
			if bytes, err := ss.ParseBytes(arg); err != nil {
				log.Printf("parse bytes failed: %v", err)
			} else {
				log.Printf("%s => %d", arg, bytes)
			}
		}
	}

	return nil
}
