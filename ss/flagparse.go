package ss

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/ver"
	"github.com/bingoohuang/ngg/yaml"
	"github.com/spf13/pflag"
)

type FlagPostProcessor interface {
	PostProcess()
}

type FlagVersionShower interface {
	VersionInfo() string
}

type FlagUsageShower interface {
	Usage() string
}

type requiredVar struct {
	name string
	p    *string
	pp   *[]string
}

type FlagParseOptions struct {
	flagName, defaultCnf string
	cnf                  *string
}

type FlagParseOptionsFn func(*FlagParseOptions)

func AutoLoadYaml(flagName, defaultCnf string) FlagParseOptionsFn {
	return func(o *FlagParseOptions) {
		o.flagName = flagName
		o.defaultCnf = defaultCnf
	}
}

func ParseFlag(a interface{}, optionFns ...FlagParseOptionsFn) {
	ParseArgs(a, os.Args, optionFns...)
}

func ParseArgs(a interface{}, args []string, optionFns ...FlagParseOptionsFn) {
	options := createOptions(optionFns)

	f := pflag.NewFlagSet(args[0], pflag.ExitOnError)
	var checkVersionShow func()
	requiredVars := make([]requiredVar, 0)

	var pprof *string

	ra := reflect.ValueOf(a).Elem()
	rt := ra.Type()
	for i := 0; i < rt.NumField(); i++ {
		fi, fv := rt.Field(i), ra.Field(i)
		if fi.PkgPath != "" { // ignore unexported
			continue
		}

		t := fi.Tag.Get
		name := t("flag")
		if name == "-" || !fv.CanAddr() {
			continue
		}

		if name == "" {
			name = ToKebab(fi.Name)
		}

		val, short, usage, required := t("val"), t("short"), t("usage"), t("required")
		p := fv.Addr().Interface()
		ft := fi.Type
		if reflect.PointerTo(ft).Implements(flagValueType) {
			pp := p.(pflag.Value)
			if val != "" {
				pp.Set(val)
			}
			f.VarP(pp, name, short, usage)
			continue
		} else if ft == timeDurationType {
			f.DurationVarP(p.(*time.Duration), name, short, ParseDuration(val), usage)
			continue
		}

		switch ft.Kind() {
		case reflect.Slice:
			switch ft.Elem().Kind() {
			case reflect.String:
				pp := p.(*[]string)
				f.StringArrayVarP(pp, name, short, []string{val}, usage)
				if required == "true" {
					requiredVars = append(requiredVars, requiredVar{name: name, pp: pp})
				}
			}
		case reflect.String:
			pp := p.(*string)
			f.StringVarP(pp, name, short, val, usage)
			if required == "true" {
				requiredVars = append(requiredVars, requiredVar{name: name, p: pp})
			}

			switch {
			case AnyOf("pprof", name, short):
				pprof = pp
			case AnyOf(options.flagName, name, short):
				options.cnf = pp
			}

		case reflect.Int:
			pp := p.(*int)
			if count := t("count"); count == "true" {
				if val != "" {
					*pp, _ = Parse[int](val)
				}
				f.CountVarP(pp, name, short, usage)
			} else {
				f.IntVarP(pp, name, short, Pick1(Parse[int](val)), usage)
			}
		case reflect.Int32:
			f.Int32VarP(p.(*int32), name, short, Pick1(Parse[int32](val)), usage)
		case reflect.Int64:
			f.Int64VarP(p.(*int64), name, short, Pick1(Parse[int64](val)), usage)
		case reflect.Uint:
			f.UintVarP(p.(*uint), name, short, Pick1(Parse[uint](val)), usage)
		case reflect.Uint32:
			f.Uint32VarP(p.(*uint32), name, short, Pick1(Parse[uint32](val)), usage)
		case reflect.Uint64:
			f.Uint64VarP(p.(*uint64), name, short, Pick1(Parse[uint64](val)), usage)
		case reflect.Bool:
			pp := p.(*bool)
			checkVersionShow = checkVersion(checkVersionShow, a, fi.Name, pp)
			f.BoolVarP(pp, name, short, Pick1(ParseBool(val)), usage)
		case reflect.Float32:
			f.Float32VarP(p.(*float32), name, short, Pick1(Parse[float32](val)), usage)
		case reflect.Float64:
			f.Float64VarP(p.(*float64), name, short, Pick1(Parse[float64](val)), usage)
		}
	}

	if u, ok := a.(FlagUsageShower); ok {
		f.Usage = func() {
			fmt.Println(strings.TrimSpace(u.Usage()))
		}
	}

	if options.cnf != nil {
		fn, sn := Split2(options.flagName, ",")
		if value, _ := FindFlag(args, fn, sn); value != "" || options.defaultCnf != "" {
			if err := loadYamlConfFile(value, options.defaultCnf, a); err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(-1)
			}
		}
	}

	// 提前到这里，实际上是为了先解析出 --conf 参数，便于下面从配置文件载入数据
	// 但是，命令行应该优先级，应该比配置文件优先级高，为了解决这个矛盾
	// 需要把 --conf 参数置为第一个参数，并且使用自定义参数的形式，在解析到改参数时，
	// 立即从对应的配置文件加载所有配置，然后再依次处理其它命令行参数
	_ = f.Parse(args[1:])

	if checkVersionShow != nil {
		checkVersionShow()
	}

	checkRequired(requiredVars, f)

	if pp, ok := a.(FlagPostProcessor); ok {
		pp.PostProcess()
	}

	if pprof != nil && *pprof != "" {
		go startPprof(*pprof)
	}
}

func FindFlag(args []string, targetNames ...string) (value string, found bool) {
	for i := 1; i < len(args); i++ {
		s := args[i]
		if len(s) < 2 || s[0] != '-' {
			continue
		}
		numMinuses := 1
		if s[1] == '-' {
			numMinuses++
			if len(s) == 2 { // "--" terminates the flags
				break
			}
		}

		name := s[numMinuses:]
		if len(name) == 0 || name[0] == '-' || name[0] == '=' { // bad flag syntax: %s"
			continue
		}
		if strings.HasPrefix(name, "test.") { // ignore go test flags
			continue
		}

		// it's a pflag. does it have an argument?
		hasValue := false
		for j := 1; j < len(name); j++ { // equals cannot be first
			if name[j] == '=' {
				value = name[j+1:]
				hasValue = true
				name = name[0:j]
				break
			}
		}

		if !AnyOf(name, targetNames...) {
			continue
		}

		// It must have a value, which might be the next argument.
		if !hasValue && i+1 < len(args) {
			// value is the next arg
			hasValue = true
			value = args[i+1]
		}

		return value, true
	}

	return "", false
}

func createOptions(fns []FlagParseOptionsFn) *FlagParseOptions {
	options := &FlagParseOptions{}
	for _, f := range fns {
		f(options)
	}

	return options
}

var (
	timeDurationType = reflect.TypeOf(time.Duration(0))
	flagValueType    = reflect.TypeOf((*pflag.Value)(nil)).Elem()
)

func checkRequired(requiredVars []requiredVar, f *pflag.FlagSet) {
	requiredMissed := 0
	for _, rv := range requiredVars {
		if rv.p != nil && *rv.p == "" || rv.pp != nil && len(*rv.pp) == 0 {
			requiredMissed++
			fmt.Printf("-%s is required\n", rv.name)
		}
	}

	if requiredMissed > 0 {
		f.Usage()
		os.Exit(1)
	}
}

func checkVersion(checker func(), arg interface{}, fiName string, bp *bool) func() {
	if checker == nil && fiName == "Version" {
		if vs, ok := arg.(FlagVersionShower); ok {
			return func() {
				if *bp {
					fmt.Println(vs.VersionInfo())
					os.Exit(0)
				}
			}
		} else {
			return func() {
				if *bp {
					fmt.Println(ver.Version())
					os.Exit(0)
				}
			}
		}
	}

	return checker
}

func startPprof(pprofAddr string) {
	pprofHostPort := pprofAddr
	parts := strings.Split(pprofHostPort, ":")
	if len(parts) == 2 && parts[0] == "" {
		pprofHostPort = fmt.Sprintf("localhost:%s", parts[1])
	}

	log.Printf("I! Starting pprof HTTP server at: http://%s/debug/pprof", pprofHostPort)
	if err := http.ListenAndServe(pprofAddr, nil); err != nil {
		log.Fatal("E! " + err.Error())
	}
}

func loadYamlConfFile(confFile, defaultConfFile string, app interface{}) error {
	if confFile == "" {
		if s, err := os.Stat(defaultConfFile); err != nil || s.IsDir() {
			return nil // not exists
		}
		confFile = defaultConfFile
	}

	data, err := os.ReadFile(confFile)
	if err != nil {
		return fmt.Errorf("read conf file %s error: %q", confFile, err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data), yaml.CustomUnmarshaler(func(t *FlagSize, b []byte) error {
		if b[0] == '"' {
			val := string(b[1 : len(b)-1])
			if v, err := ParseBytes(val); err != nil {
				return err
			} else {
				*t = FlagSize(v)
			}
		}
		val, err := Parse[uint64](string(b))
		if err != nil {
			return err
		}

		*t = FlagSize(val)
		return nil
	}))

	if err := decoder.Decode(app); err != nil {
		return fmt.Errorf("decode conf file %s error:%q", confFile, err)
	}

	return nil
}
