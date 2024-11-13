module github.com/bingoohuang/ngg/kt

go 1.22.0

toolchain go1.23.3

replace (
	github.com/bingoohuang/ngg/mapstruct => ../mapstruct
	github.com/bingoohuang/ngg/ss => ../ss
)

require (
	github.com/AndrewBurian/eventsource v2.1.0+incompatible
	github.com/IBM/sarama v1.43.3
	github.com/bingoohuang/ngg/daemon v0.0.0-20241113020638-78201765d5cb
	github.com/bingoohuang/ngg/jj v0.0.0-20241113020638-78201765d5cb
	github.com/bingoohuang/ngg/mapstruct v0.0.0-00010101000000-000000000000
	github.com/bingoohuang/ngg/ss v0.0.0-20241113020638-78201765d5cb
	github.com/bingoohuang/ngg/ver v0.0.0-20241113020638-78201765d5cb
	github.com/elliotchance/pie/v2 v2.9.0
	github.com/joho/godotenv v1.5.1
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.35.1
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475
	github.com/samber/lo v1.47.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.9.0
	github.com/vmihailenco/tagparser/v2 v2.0.0
	golang.org/x/term v0.26.0
)

require (
	github.com/Pallinder/go-randomdata v1.2.0 // indirect
	github.com/araddon/dateparse v0.0.0-20210429162001-6b43995a97de // indirect
	github.com/bingoohuang/ngg/q v0.0.0-20241113020638-78201765d5cb // indirect
	github.com/bingoohuang/ngg/tick v0.0.0-20241113020638-78201765d5cb // indirect
	github.com/bingoohuang/ngg/tsid v0.0.0-20241113020638-78201765d5cb // indirect
	github.com/bingoohuang/ngg/yaml v0.0.0-20241113020638-78201765d5cb // indirect
	github.com/brianvoe/gofakeit/v6 v6.28.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/eapache/go-resiliency v1.7.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/howeyc/gopass v0.0.0-20210920133722-c8aef6fb66ef // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/nxadm/tail v1.4.11 // indirect
	github.com/pbnjay/pixfont v0.0.0-20200714042608-33b744692567 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/sagikazarmark/locafero v0.6.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	github.com/sevlyar/go-daemon v0.1.6 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/spf13/viper v1.19.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.29.0 // indirect
	golang.org/x/exp v0.0.0-20241108190413-2d47ceb2692f // indirect
	golang.org/x/net v0.31.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	golang.org/x/text v0.20.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
