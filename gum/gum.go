package gum

import (
	"github.com/alecthomas/kong"
	"github.com/bingoohuang/ngg/gum/choose"
	"github.com/bingoohuang/ngg/gum/confirm"
	"github.com/bingoohuang/ngg/gum/input"
)

var kongVars = kong.Vars{
	"defaultHeight":           "0",
	"defaultWidth":            "0",
	"defaultAlign":            "left",
	"defaultBorder":           "none",
	"defaultBorderForeground": "",
	"defaultBorderBackground": "",
	"defaultBackground":       "",
	"defaultForeground":       "",
	"defaultMargin":           "0 0",
	"defaultPadding":          "0 0",
	"defaultUnderline":        "false",
	"defaultBold":             "false",
	"defaultFaint":            "false",
	"defaultItalic":           "false",
	"defaultStrikethrough":    "false",
}

func Confirm(prompt string) (bool, error) {
	option := &confirm.Options{}
	KongParse(option, kongVars)
	option.Prompt = prompt
	return option.Run()
}

func Choose(header string, options []string, limit int) ([]string, error) {
	option := &choose.Options{}
	KongParse(option, kongVars)
	option.Header = header
	option.Options = options
	option.Limit = limit
	return option.Run()
}

func Input(prompt, placeholder string) (string, error) {
	option := &input.Options{}
	KongParse(option, kongVars)
	option.Prompt = prompt
	option.Placeholder = placeholder
	return option.Run()
}

// KongParse constructs a new parser and parses the default command-line.
func KongParse(cli interface{}, options ...kong.Option) *kong.Context {
	parser, err := kong.New(cli, options...)
	if err != nil {
		panic(err)
	}
	ctx, err := parser.Parse(nil)
	parser.FatalIfErrorf(err)
	return ctx
}
