module github.com/bingoohuang/ngg/ggt

go 1.23.0

toolchain go1.23.1

replace (
	github.com/bingoohuang/ngg/godbtest => ../godbtest
	github.com/bingoohuang/ngg/gossh => ../gossh
	github.com/bingoohuang/ngg/jj => ../jj
	github.com/bingoohuang/ngg/ss => ../ss
	github.com/fatedier/frp => github.com/bingoohuang/frp v0.0.0-20240921114209-e8b9689f39d9
	github.com/emmansun/gmsm => ./gmsm
)

require (
	gitee.com/Trisia/gotlcp v1.3.23
	github.com/AlecAivazis/survey/v2 v2.3.7
	github.com/BurntSushi/toml v1.4.0
	github.com/atotto/clipboard v0.1.4
	github.com/bingoohuang/ngg/cmd v0.0.0-20240925004800-51db94c4a057
	github.com/bingoohuang/ngg/daemon v0.0.0-20240925004800-51db94c4a057
	github.com/bingoohuang/ngg/gnet v0.0.0-20240925004800-51db94c4a057
	github.com/bingoohuang/ngg/godbtest v0.0.0-20240925004800-51db94c4a057
	github.com/bingoohuang/ngg/gossh v0.0.0-20240925004800-51db94c4a057
	github.com/bingoohuang/ngg/gum v0.0.0-20240925004800-51db94c4a057
	github.com/bingoohuang/ngg/jj v0.0.0-20240925004800-51db94c4a057
	github.com/bingoohuang/ngg/rotatefile v0.0.0-20240925004800-51db94c4a057
	github.com/bingoohuang/ngg/ss v0.0.0-20240925004800-51db94c4a057
	github.com/bingoohuang/ngg/tick v0.0.0-20240925004800-51db94c4a057
	github.com/bingoohuang/ngg/ver v0.0.0-20240925004800-51db94c4a057
	github.com/bingoohuang/ngg/yaml v0.0.0-20240925004800-51db94c4a057
	github.com/cespare/xxhash/v2 v2.3.0
	github.com/chzyer/readline v1.5.1
	github.com/cloudwego/hertz v0.9.3
	github.com/deatil/go-cryptobin v1.0.4011
	github.com/dustin/go-humanize v1.0.1
	github.com/emmansun/gmsm v0.28.0
	github.com/expr-lang/expr v1.16.9
	github.com/fatedier/frp v0.60.0
	github.com/fatih/color v1.17.0
	github.com/glebarez/sqlite v1.11.0
	github.com/hertz-contrib/gzip v0.0.3
	github.com/imroc/req/v3 v3.46.1
	github.com/jedib0t/go-pretty/v6 v6.5.9
	github.com/joho/godotenv v1.5.1
	github.com/k0kubun/go-ansi v0.0.0-20180517002512-3bf9e2903213
	github.com/klauspost/cpuid/v2 v2.2.8
	github.com/mattn/go-isatty v0.0.20
	github.com/minio/sio v0.4.1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-ps v1.0.0
	github.com/pion/stun/v2 v2.0.0
	github.com/redis/go-redis/v9 v9.6.1
	github.com/samber/lo v1.47.0
	github.com/schollz/pake/v3 v3.0.5
	github.com/schollz/progressbar/v3 v3.16.0
	github.com/segmentio/ksuid v1.0.4
	github.com/shirou/gopsutil/v4 v4.24.8
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.0.1010
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse v1.0.1010
	github.com/vthiery/retry v0.1.0
	github.com/zeebo/blake3 v0.2.4
	go.uber.org/atomic v1.11.0
	go.uber.org/multierr v1.11.0
	golang.org/x/crypto v0.27.0
	golang.org/x/term v0.24.0
	golang.org/x/time v0.6.0
	gorm.io/gorm v1.25.12
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	gitee.com/chunanyong/dm v1.8.16 // indirect
	gitee.com/opengauss/openGauss-connector-go-pq v1.0.4 // indirect
	github.com/Azure/go-ntlmssp v0.0.0-20221128193559-754e69321358 // indirect
	github.com/Pallinder/go-randomdata v1.2.0 // indirect
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/alecthomas/chroma v0.10.0 // indirect
	github.com/alecthomas/kong v1.2.1 // indirect
	github.com/andeya/ameda v1.5.3 // indirect
	github.com/andeya/goutil v1.0.1 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/bingoohuang/ngg/q v0.0.0-20240925004800-51db94c4a057 // indirect
	github.com/bingoohuang/ngg/sqlparser v0.0.0-20240925004800-51db94c4a057 // indirect
	github.com/bingoohuang/ngg/tsid v0.0.0-20240925004800-51db94c4a057 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/brianvoe/gofakeit/v6 v6.28.0 // indirect
	github.com/bytedance/go-tagexpr/v2 v2.9.11 // indirect
	github.com/bytedance/gopkg v0.1.1 // indirect
	github.com/bytedance/sonic v1.12.3 // indirect
	github.com/bytedance/sonic/loader v0.2.0 // indirect
	github.com/catppuccin/go v0.2.0 // indirect
	github.com/charmbracelet/bubbles v0.20.0 // indirect
	github.com/charmbracelet/bubbletea v1.1.1 // indirect
	github.com/charmbracelet/gum v0.14.5 // indirect
	github.com/charmbracelet/huh v0.6.0 // indirect
	github.com/charmbracelet/lipgloss v0.13.0 // indirect
	github.com/charmbracelet/x/ansi v0.3.2 // indirect
	github.com/charmbracelet/x/exp/strings v0.0.0-20240919170804-a4978c8e603a // indirect
	github.com/charmbracelet/x/term v0.2.0 // indirect
	github.com/cheggaaa/pb/v3 v3.1.5 // indirect
	github.com/cloudflare/circl v1.4.0 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/cloudwego/netpoll v0.6.4 // indirect
	github.com/creasty/defaults v1.8.0 // indirect
	github.com/denisenkom/go-mssqldb v0.12.3 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dlclark/regexp2 v1.11.4 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/fatedier/golib v0.5.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/glebarez/go-sqlite v1.22.0 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/gobars/cmd v0.0.0-20210215022658-cd78beda9673 // indirect
	github.com/gohxs/readline v0.0.0-20171011095936-a780388e6e7c // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/pprof v0.0.0-20240910150728-a0b0bb1d4134 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/h2non/filetype v1.1.3 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/howeyc/gopass v0.0.0-20210920133722-c8aef6fb66ef // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.1 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.17.10 // indirect
	github.com/klauspost/reedsolomon v1.12.4 // indirect
	github.com/kortschak/goroutine v1.1.2 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20240909124753-873cd0166683 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mattn/go-shellwords v1.0.12 // indirect
	github.com/mattn/go-sqlite3 v1.14.23 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/termenv v0.15.3-0.20240618155329-98d742f6907a // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/nyaruka/phonenumbers v1.4.0 // indirect
	github.com/onsi/ginkgo/v2 v2.20.2 // indirect
	github.com/orcaman/concurrent-map/v2 v2.0.1 // indirect
	github.com/pbnjay/pixfont v0.0.0-20200714042608-33b744692567 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pion/dtls/v2 v2.2.12 // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/transport/v2 v2.2.10 // indirect
	github.com/pion/transport/v3 v3.0.7 // indirect
	github.com/pires/go-proxyproto v0.7.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/sftp v1.13.6 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/quic-go/qpack v0.5.1 // indirect
	github.com/quic-go/quic-go v0.47.0 // indirect
	github.com/refraction-networking/utls v1.6.7 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/sagikazarmark/locafero v0.6.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sevlyar/go-daemon v0.1.6 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/sijms/go-ora/v2 v2.8.21 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/spf13/viper v1.19.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/templexxx/cpu v0.1.1 // indirect
	github.com/templexxx/xorsimd v0.4.3 // indirect
	github.com/tidwall/gjson v1.17.3 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	github.com/tklauser/go-sysconf v0.3.14 // indirect
	github.com/tklauser/numcpus v0.8.0 // indirect
	github.com/tscholl2/siec v0.0.0-20240310163802-c2c6f6198406 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/wlynxg/anet v0.0.4 // indirect
	github.com/xo/dburl v0.23.2 // indirect
	github.com/xtaci/kcp-go/v5 v5.6.17 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/assert v1.3.0 // indirect
	go.uber.org/mock v0.4.0 // indirect
	golang.org/x/arch v0.10.0 // indirect
	golang.org/x/exp v0.0.0-20240909161429-701f63a606c0 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	golang.org/x/tools v0.25.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/protobuf v1.34.2 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apimachinery v0.31.1 // indirect
	k8s.io/utils v0.0.0-20240921022957-49e7df575cb6 // indirect
	modernc.org/libc v1.61.0 // indirect
	modernc.org/mathutil v1.6.0 // indirect
	modernc.org/memory v1.8.0 // indirect
	modernc.org/sqlite v1.33.1 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)
