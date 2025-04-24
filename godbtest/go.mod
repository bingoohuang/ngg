module github.com/bingoohuang/ngg/godbtest

go 1.24

toolchain go1.24.2

replace (
	github.com/bingoohuang/ngg/cmd => ../cmd
	github.com/bingoohuang/ngg/ggt => ../../poc/ggt
	github.com/bingoohuang/ngg/jj => ../jj
	github.com/bingoohuang/ngg/ss => ../ss
)

require (
	gitee.com/chunanyong/dm v1.8.19
	gitee.com/opengauss/openGauss-connector-go-pq v1.0.7
	github.com/AlecAivazis/survey/v2 v2.3.7
	github.com/Pallinder/go-randomdata v1.2.0
	github.com/alecthomas/chroma v0.10.0
	github.com/bingoohuang/ngg/cmd v0.0.0-20250404002712-8ac671fa9708
	github.com/bingoohuang/ngg/ggt v0.0.0-20250227024459-8f5fb181c16c
	github.com/bingoohuang/ngg/jj v0.0.0-20250404002712-8ac671fa9708
	github.com/bingoohuang/ngg/pp v0.0.0-20250404002712-8ac671fa9708
	github.com/bingoohuang/ngg/sqlparser v0.0.0-20250404002712-8ac671fa9708
	github.com/bingoohuang/ngg/ss v0.0.0-20250404002712-8ac671fa9708
	github.com/bingoohuang/ngg/tick v0.0.0-20250404002712-8ac671fa9708
	github.com/bingoohuang/ngg/ver v0.0.0-20250404002712-8ac671fa9708
	github.com/cespare/xxhash/v2 v2.3.0
	github.com/cheggaaa/pb/v3 v3.1.7
	github.com/creasty/defaults v1.8.0
	github.com/deatil/go-cryptobin v1.1.1002
	github.com/denisenkom/go-mssqldb v0.12.3
	github.com/emirpasic/gods v1.18.1
	github.com/go-sql-driver/mysql v1.9.2
	github.com/gohxs/readline v0.0.0-20171011095936-a780388e6e7c
	github.com/google/go-cmp v0.7.0
	github.com/h2non/filetype v1.1.3
	github.com/imroc/req/v3 v3.50.0
	github.com/jackc/pgx/v5 v5.7.4
	github.com/jedib0t/go-pretty/v6 v6.6.7
	github.com/lib/pq v1.10.9
	github.com/mattn/go-shellwords v1.0.12
	github.com/mattn/go-sqlite3 v1.14.28
	github.com/orcaman/concurrent-map/v2 v2.0.1
	github.com/samber/lo v1.49.1
	github.com/sijms/go-ora/v2 v2.8.24
	github.com/spf13/cobra v1.9.1
	github.com/spf13/pflag v1.0.6
	github.com/stretchr/testify v1.10.0
	github.com/xo/dburl v0.23.7
	go.uber.org/atomic v1.11.0
	go.uber.org/multierr v1.11.0
	golang.org/x/net v0.39.0
	gopkg.in/yaml.v3 v3.0.1
	modernc.org/sqlite v1.37.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/bingoohuang/ngg/tsid v0.0.0-20250404002712-8ac671fa9708 // indirect
	github.com/brianvoe/gofakeit/v6 v6.28.0 // indirect
	github.com/chzyer/test v1.0.0 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/goccy/go-yaml v1.17.1 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/pprof v0.0.0-20250423184734-337e5dd93bb4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/howeyc/gopass v0.0.0-20210920133722-c8aef6fb66ef // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/onsi/ginkgo/v2 v2.23.4 // indirect
	github.com/pbnjay/pixfont v0.0.0-20200714042608-33b744692567 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/quic-go/qpack v0.5.1 // indirect
	github.com/quic-go/quic-go v0.51.0 // indirect
	github.com/refraction-networking/utls v1.7.1 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/sagikazarmark/locafero v0.9.0 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.14.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/viper v1.20.1 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	go.uber.org/mock v0.5.1 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/term v0.31.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	golang.org/x/tools v0.32.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	modernc.org/libc v1.63.0 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.10.0 // indirect
)
