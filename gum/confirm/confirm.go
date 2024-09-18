package confirm

import (
	"fmt"
	"time"

	"github.com/charmbracelet/gum/style"
	"github.com/charmbracelet/huh"
)

// Options is the customization options for the confirm command.
type Options struct {
	Default     bool   `help:"Default confirmation action" default:"true"`
	Affirmative string `help:"The title of the affirmative action" default:"Yes"`
	Negative    string `help:"The title of the negative action" default:"No"`
	Prompt      string `arg:"" help:"Prompt to display." default:"Are you sure?"`
	//nolint:staticcheck
	PromptStyle style.Styles `embed:"" prefix:"prompt." help:"The style of the prompt" set:"defaultMargin=0 0 0 1" set:"defaultForeground=#7571F9" set:"defaultBold=true" envprefix:"GUM_CONFIRM_PROMPT_"`
	//nolint:staticcheck
	SelectedStyle style.Styles `embed:"" prefix:"selected." help:"The style of the selected action" set:"defaultBackground=212" set:"defaultForeground=230" set:"defaultPadding=0 3" set:"defaultMargin=0 1" envprefix:"GUM_CONFIRM_SELECTED_"`
	//nolint:staticcheck
	UnselectedStyle style.Styles  `embed:"" prefix:"unselected." help:"The style of the unselected action" set:"defaultBackground=235" set:"defaultForeground=254" set:"defaultPadding=0 3" set:"defaultMargin=0 1" envprefix:"GUM_CONFIRM_UNSELECTED_"`
	ShowHelp        bool          `help:"Show help key binds" negatable:"" default:"true" env:"GUM_CONFIRM_SHOW_HELP"`
	Timeout         time.Duration `help:"Timeout until confirm returns selected value or default if provided" default:"0" env:"GUM_CONFIRM_TIMEOUT"`
}

// Run provides a shell script interface for prompting a user to confirm an
// action with an affirmative or negative answer.
func (o Options) Run() (bool, error) {
	theme := huh.ThemeCharm()
	theme.Focused.Title = o.PromptStyle.ToLipgloss()
	theme.Focused.FocusedButton = o.SelectedStyle.ToLipgloss()
	theme.Focused.BlurredButton = o.UnselectedStyle.ToLipgloss()

	choice := o.Default

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Affirmative(o.Affirmative).
				Negative(o.Negative).
				Title(o.Prompt).
				Value(&choice),
		),
	).
		WithTheme(theme).
		WithShowHelp(o.ShowHelp).
		Run()

	if err != nil {
		return false, fmt.Errorf("unable to run confirm: %w", err)
	}

	return choice, nil
}
