module github.com/bingoohuang/ngg/ggt

go 1.23.0

toolchain go1.23.1

replace (
	github.com/bingoohuang/ngg/gnet => ../gnet
	github.com/bingoohuang/ngg/gossh => ../gossh
	github.com/bingoohuang/ngg/gum => ../gum
	github.com/bingoohuang/ngg/jj => ../jj
	github.com/bingoohuang/ngg/sqlparser => ../sqlparser
	github.com/bingoohuang/ngg/ss => ../ss
	github.com/bingoohuang/ngg/ver => ../ver
	github.com/emmansun/gmsm => ./gmsm
)

require (
	github.com/bingoohuang/ngg/daemon v0.0.0-20241103104547-2d1bf667deef
	github.com/bingoohuang/ngg/gum v0.0.0-20241103104547-2d1bf667deef
	github.com/bingoohuang/ngg/rotatefile v0.0.0-20241103104547-2d1bf667deef
	github.com/bingoohuang/ngg/ss v0.0.0-20241103104547-2d1bf667deef
	github.com/bingoohuang/ngg/ver v0.0.0-20241103104547-2d1bf667deef
	github.com/joho/godotenv v1.5.1
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
)

require (
	github.com/alecthomas/kong v1.4.0 // indirect
	github.com/atotto/clipboard v0.1.4 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/bingoohuang/ngg/q v0.0.0-20241103104547-2d1bf667deef // indirect
	github.com/bingoohuang/ngg/yaml v0.0.0-20241103104547-2d1bf667deef // indirect
	github.com/catppuccin/go v0.2.0 // indirect
	github.com/charmbracelet/bubbles v0.20.0 // indirect
	github.com/charmbracelet/bubbletea v1.2.4 // indirect
	github.com/charmbracelet/gum v0.14.5 // indirect
	github.com/charmbracelet/huh v0.6.0 // indirect
	github.com/charmbracelet/lipgloss v1.0.0 // indirect
	github.com/charmbracelet/x/ansi v0.5.2 // indirect
	github.com/charmbracelet/x/exp/strings v0.0.0-20241122161412-4559bf4d941d // indirect
	github.com/charmbracelet/x/term v0.2.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/howeyc/gopass v0.0.0-20210920133722-c8aef6fb66ef // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/kortschak/goroutine v1.1.2 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/termenv v0.15.3-0.20240618155329-98d742f6907a // indirect
	github.com/pbnjay/pixfont v0.0.0-20200714042608-33b744692567 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/sagikazarmark/locafero v0.6.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sevlyar/go-daemon v0.1.6 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/spf13/viper v1.19.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/exp v0.0.0-20241009180824-f66d83c29e7c // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sync v0.9.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	golang.org/x/term v0.25.0 // indirect
	golang.org/x/text v0.20.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
