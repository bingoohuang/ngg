package input

import (
	"os"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/gum/style"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Options are the customization options for the input.
type Options struct {
	Placeholder      string        `help:"Placeholder value" default:"Type something..." env:"GUM_INPUT_PLACEHOLDER"`
	Prompt           string        `help:"Prompt to display" default:"> " env:"GUM_INPUT_PROMPT"`
	PromptStyle      style.Styles  `embed:"" prefix:"prompt." envprefix:"GUM_INPUT_PROMPT_"`
	PlaceholderStyle style.Styles  `embed:"" prefix:"placeholder." set:"defaultForeground=240" envprefix:"GUM_INPUT_PLACEHOLDER_"`
	CursorStyle      style.Styles  `embed:"" prefix:"cursor." set:"defaultForeground=212" envprefix:"GUM_INPUT_CURSOR_"`
	CursorMode       string        `prefix:"cursor." name:"mode" help:"Cursor mode" default:"blink" enum:"blink,hide,static" env:"GUM_INPUT_CURSOR_MODE"`
	Value            string        `help:"Initial value (can also be passed via stdin)" default:""`
	CharLimit        int           `help:"Maximum value length (0 for no limit)" default:"400"`
	Width            int           `help:"Input width (0 for terminal width)" default:"0" env:"GUM_INPUT_WIDTH"`
	Password         bool          `help:"Mask input characters" default:"false"`
	ShowHelp         bool          `help:"Show help keybinds" default:"true" negatable:"true" env:"GUM_INPUT_SHOW_HELP"`
	Header           string        `help:"Header value" default:"" env:"GUM_INPUT_HEADER"`
	HeaderStyle      style.Styles  `embed:"" prefix:"header." set:"defaultForeground=240" envprefix:"GUM_INPUT_HEADER_"`
	Timeout          time.Duration `help:"Timeout until input aborts" default:"0" env:"GUM_INPUT_TIMEOUT"`
}

// Run provides a shell script interface for the text input bubble.
// https://github.com/charmbracelet/bubbles/textinput
func (o Options) Run() (string, error) {
	value := o.Value

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
		WithShowHelp(o.ShowHelp).
		WithProgramOptions(tea.WithOutput(os.Stderr)).
		Run()
	if err != nil {
		return "", err
	}

	return value, nil
}
