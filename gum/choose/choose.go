package choose

import (
	"time"

	"github.com/charmbracelet/gum/style"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Options is the customization options for the choose command.
type Options struct {
	Options           []string      `arg:"" optional:"" help:"Options to choose from."`
	Limit             int           `help:"Maximum number of options to pick" default:"1" group:"Selection"`
	NoLimit           bool          `help:"Pick unlimited number of options (ignores limit)" group:"Selection"`
	Ordered           bool          `help:"Maintain the order of the selected options" env:"GUM_CHOOSE_ORDERED"`
	Height            int           `help:"Height of the list" default:"0" env:"GUM_CHOOSE_HEIGHT"`
	Cursor            string        `help:"Prefix to show on item that corresponds to the cursor position" default:"> " env:"GUM_CHOOSE_CURSOR"`
	ShowHelp          bool          `help:"Show help keybinds" default:"true" negatable:"true" env:"GUM_CHOOSE_SHOW_HELP"`
	Header            string        `help:"Header value" default:"Choose:" env:"GUM_CHOOSE_HEADER"`
	CursorPrefix      string        `help:"Prefix to show on the cursor item (hidden if limit is 1)" default:"• " env:"GUM_CHOOSE_CURSOR_PREFIX"`
	SelectedPrefix    string        `help:"Prefix to show on selected items (hidden if limit is 1)" default:"✓ " env:"GUM_CHOOSE_SELECTED_PREFIX"`
	UnselectedPrefix  string        `help:"Prefix to show on unselected items (hidden if limit is 1)" default:"• " env:"GUM_CHOOSE_UNSELECTED_PREFIX"`
	Selected          []string      `help:"Options that should start as selected" default:"" env:"GUM_CHOOSE_SELECTED"`
	SelectIfOne       bool          `help:"Select the given option if there is only one" group:"Selection"`
	CursorStyle       style.Styles  `embed:"" prefix:"cursor." set:"defaultForeground=212" envprefix:"GUM_CHOOSE_CURSOR_"`
	HeaderStyle       style.Styles  `embed:"" prefix:"header." set:"defaultForeground=99" envprefix:"GUM_CHOOSE_HEADER_"`
	ItemStyle         style.Styles  `embed:"" prefix:"item." hidden:"" envprefix:"GUM_CHOOSE_ITEM_"`
	SelectedItemStyle style.Styles  `embed:"" prefix:"selected." set:"defaultForeground=212" envprefix:"GUM_CHOOSE_SELECTED_"`
	Timeout           time.Duration `help:"Timeout until choose returns selected element" default:"0" env:"GUM_CCHOOSE_TIMEOUT"` // including timeout command options [Timeout,...]
}

const widthBuffer = 2

// Run provides a shell script interface for choosing between different through
// options.
func (o Options) Run() ([]string, error) {
	if len(o.Options) <= 0 {
		panic("options is empty")
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
			Run()
		if err != nil {
			return nil, err
		}
		if len(choices) > 0 {
			return choices, nil
		}
		return nil, nil
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
		WithShowHelp(o.ShowHelp).
		Run()
	if err != nil {
		return nil, err
	}

	return []string{choice}, nil
}

func widest(options []string) int {
	var max int
	for _, o := range options {
		w := lipgloss.Width(o)
		if w > max {
			max = w
		}
	}
	return max
}
