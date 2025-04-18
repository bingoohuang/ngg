package confirm

import (
	"github.com/bingoohuang/ngg/gum/internal/exit"
	"github.com/charmbracelet/huh"
)

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
		WithTimeout(o.Timeout).
		WithTheme(theme).
		WithShowHelp(o.ShowHelp).
		Run()

	if err != nil {
		return false, exit.Handle(err, o.Timeout)
	}

	return choice, nil
}
