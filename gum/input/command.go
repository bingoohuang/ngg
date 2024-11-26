package input

import (
	"os"

	"github.com/bingoohuang/ngg/gum/internal/exit"
	"github.com/bingoohuang/ngg/gum/internal/stdin"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Run provides a shell script interface for the text input bubble.
// https://github.com/charmbracelet/bubbles/textinput
func (o Options) Run() (string, error) {
	var value string
	if o.Value != "" {
		value = o.Value
	} else if in, _ := stdin.Read(); in != "" {
		value = in
	}

	theme := huh.ThemeCharm()
	theme.Focused.Base = lipgloss.NewStyle()
	theme.Focused.TextInput.Cursor = o.CursorStyle.ToLipgloss()
	theme.Focused.TextInput.Placeholder = o.PlaceholderStyle.ToLipgloss()
	theme.Focused.TextInput.Prompt = o.PromptStyle.ToLipgloss()
	theme.Focused.Title = o.HeaderStyle.ToLipgloss()

	// Keep input keymap backwards compatible
	keymap := huh.NewDefaultKeyMap()
	keymap.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"))

	echoMode := huh.EchoModeNormal
	if o.Password {
		echoMode = huh.EchoModePassword
	}

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Prompt(o.Prompt).
				Placeholder(o.Placeholder).
				CharLimit(o.CharLimit).
				EchoMode(echoMode).
				Title(o.Header).
				Value(&value),
		),
	).
		WithShowHelp(false).
		WithWidth(o.Width).
		WithTheme(theme).
		WithKeyMap(keymap).
		WithTimeout(o.Timeout).
		WithShowHelp(o.ShowHelp).
		WithProgramOptions(tea.WithOutput(os.Stderr)).
		Run()
	if err != nil {
		return "", exit.Handle(err, o.Timeout)
	}

	return value, nil
}
