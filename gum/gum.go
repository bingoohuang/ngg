package gum

import (
	"time"

	"github.com/alecthomas/kong"
	"github.com/bingoohuang/ngg/gum/choose"
	"github.com/bingoohuang/ngg/gum/confirm"
	"github.com/bingoohuang/ngg/gum/input"
)

var KongVars = kong.Vars{
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

func ConfirmTimeout(timeout time.Duration) func(*confirm.Options) {
	return func(o *confirm.Options) { o.Timeout = timeout }
}

func ConfirmDefault(defaultValue bool) func(*confirm.Options) {
	return func(o *confirm.Options) { o.Default = defaultValue }
}

func Confirm(prompt string, optionsFn ...func(*confirm.Options) ([]string, error)) (bool, error) {
	option := &confirm.Options{}
	KongParse(option, KongVars)
	option.Prompt = prompt

	for _, fn := range optionsFn {
		fn(option)
	}

	return option.Run()
}

func ChooseTimeout(timeout time.Duration) func(*choose.Options) {
	return func(o *choose.Options) { o.Timeout = timeout }
}

func ChooseTimeoutValues(timeoutValues []string) func(*choose.Options) {
	return func(o *choose.Options) { o.TimeoutValues = timeoutValues }
}

func ChooseLimit(limit int) func(*choose.Options) {
	return func(o *choose.Options) { o.Limit = limit }
}
func ChooseHeader(header string) func(*choose.Options) {
	return func(o *choose.Options) { o.Header = header }
}

func Choose(options []string, optionsFn ...func(*choose.Options)) ([]string, error) {
	option := &choose.Options{}
	KongParse(option, KongVars)

	option.Options = options
	for _, fn := range optionsFn {
		fn(option)
	}
	return option.Run()
}

func InputTimeout(timeout time.Duration) func(*input.Options) {
	return func(o *input.Options) { o.Timeout = timeout }
}

func InputPlaceholder(placehold string) func(*input.Options) {
	return func(o *input.Options) { o.Placeholder = placehold }
}

func Input(prompt string, optionsFn ...func(*input.Options)) (string, error) {
	option := &input.Options{}
	KongParse(option, KongVars)
	option.Prompt = prompt

	for _, fn := range optionsFn {
		fn(option)
	}

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
