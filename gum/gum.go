package gum

import (
	"github.com/alecthomas/kong"
	"github.com/bingoohuang/ngg/gum/choose"
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

func Choose(options []string, limit int) ([]string, error) {
	option := &choose.Options{}
	kong.Parse(option, kongVars)
	option.Options = options
	option.Limit = limit
	return option.Run()
}

func Input(prompt, placeholder string) (string, error) {
	option := &input.Options{}
	kong.Parse(option, kongVars)
	option.Prompt = prompt
	option.Placeholder = placeholder
	return option.Run()
}
