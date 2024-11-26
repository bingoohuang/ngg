package choose

import (
	"errors"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"

	"github.com/bingoohuang/ngg/gum/internal/exit"
	"github.com/bingoohuang/ngg/gum/internal/stdin"
)

const widthBuffer = 2

// Run provides a shell script interface for choosing between different through
// options.
func (o Options) Run() ([]string, error) {
	if len(o.Options) <= 0 {
		input, _ := stdin.Read()
		if input == "" {
			return nil, errors.New("no options provided, see `gum choose --help`")
		}
		o.Options = strings.Split(input, "\n")
	}

	if o.SelectIfOne && len(o.Options) == 1 {
		return o.Options, nil
	}

	theme := huh.ThemeCharm()
	options := huh.NewOptions(o.Options...)

	theme.Focused.Base = lipgloss.NewStyle()
	theme.Focused.Title = o.HeaderStyle.ToLipgloss()
	theme.Focused.SelectSelector = o.CursorStyle.ToLipgloss().SetString(o.Cursor)
	theme.Focused.MultiSelectSelector = o.CursorStyle.ToLipgloss().SetString(o.Cursor)
	theme.Focused.SelectedOption = o.SelectedItemStyle.ToLipgloss()
	theme.Focused.UnselectedOption = o.ItemStyle.ToLipgloss()
	theme.Focused.SelectedPrefix = o.SelectedItemStyle.ToLipgloss().SetString(o.SelectedPrefix)
	theme.Focused.UnselectedPrefix = o.ItemStyle.ToLipgloss().SetString(o.UnselectedPrefix)

	for _, s := range o.Selected {
		for i, opt := range options {
			if s == opt.Key || s == opt.Value {
				options[i] = opt.Selected(true)
			}
		}
	}

	width := max(widest(o.Options)+
		max(lipgloss.Width(o.SelectedPrefix)+lipgloss.Width(o.UnselectedPrefix))+
		lipgloss.Width(o.Cursor)+1, lipgloss.Width(o.Header)+widthBuffer)

	if o.NoLimit {
		o.Limit = 0
	}

	if o.Limit > 1 || o.NoLimit {
		var choices []string

		field := huh.NewMultiSelect[string]().
			Options(options...).
			Title(o.Header).
			Height(o.Height).
			Limit(o.Limit).
			Value(&choices)

		form := huh.NewForm(huh.NewGroup(field))

		err := form.
			WithWidth(width).
			WithShowHelp(o.ShowHelp).
			WithTheme(theme).
			WithTimeout(o.Timeout).
			Run()
		if err != nil {
			if len(o.TimeoutValues) > 0 && errors.Is(err, huh.ErrTimeout) {
				return o.TimeoutValues, nil
			}

			return nil, exit.Handle(err, o.Timeout)
		}
		if len(choices) > 0 {
			if !term.IsTerminal(os.Stdout.Fd()) {
				for i, s := range choices {
					choices[i] = ansi.Strip(s)
				}
			}
		}
		return choices, nil
	}

	var choice string

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Options(options...).
				Title(o.Header).
				Height(o.Height).
				Value(&choice),
		),
	).
		WithWidth(width).
		WithTheme(theme).
		WithTimeout(o.Timeout).
		WithShowHelp(o.ShowHelp).
		Run()
	if err != nil {
		if len(o.TimeoutValues) > 0 && errors.Is(err, huh.ErrTimeout) {
			return o.TimeoutValues, nil
		}

		return nil, exit.Handle(err, o.Timeout)
	}

	if term.IsTerminal(os.Stdout.Fd()) {
		return []string{choice}, nil
	}

	return []string{ansi.Strip(choice)}, nil
}

func widest(options []string) int {
	var maxw int
	for _, o := range options {
		w := lipgloss.Width(o)
		if w > maxw {
			maxw = w
		}
	}
	return maxw
}
