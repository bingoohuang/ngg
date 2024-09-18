package root

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"time"

	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/ver"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type RootCmd struct {
	*cobra.Command
}

func create() *RootCmd {
	// rootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:   "ggt",
		Short: "golang tools",
	}

	rootCmd.Version = "version"
	rootCmd.SetVersionTemplate(ver.Version() + "\n")

	r := &RootCmd{Command: rootCmd}
	return r
}

var Cmd = create()

func Run() {
	if err := Cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

func InitFlags(f any, p *pflag.FlagSet) error {
	ptrVal := reflect.ValueOf(f)
	structVal := ptrVal.Elem()
	structType := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		field := structType.Field(i)
		tags, err := ss.ParseStructTags(string(field.Tag))
		if err != nil {
			return err
		}

		if kong, _ := tags.Get("kong"); kong != nil && kong.Raw == "-" {
			continue
		}

		short := ""
		if shortTag, _ := tags.Get("short"); shortTag != nil {
			short = shortTag.Raw
		}
		usage := ""
		if usageTag, _ := tags.Get("help"); usageTag != nil {
			usage = usageTag.Raw
		}
		defaultVal := ""
		if defaultTag, _ := tags.Get("default"); defaultTag != nil {
			defaultVal = defaultTag.Raw
		}
		if defaultVal == "" {
			if t, _ := tags.Get("env"); t != nil {
				defaultVal = os.Getenv(t.Raw)
			}
		}

		pp := structVal.Field(i).Addr().Interface()
		if pf, ok := pp.(pflag.Value); ok {
			if defaultVal != "" {
				pf.Set(defaultVal)
			}
			p.VarP(pf, ss.ToSnake(field.Name), short, usage)
			continue
		}

		if field.Type == reflect.TypeOf(time.Duration(0)) {
			var curDefault time.Duration
			if defaultVal != "" {
				if val, _, err := ss.ParseDur(defaultVal); err != nil {
					return fmt.Errorf("parse duration %s: %w", defaultVal, err)
				} else {
					curDefault = val
				}
			}
			p.DurationVarP(pp.(*time.Duration), ss.ToSnake(field.Name), short, curDefault, usage)
			continue
		}

		switch field.Type.Kind() {
		case reflect.String:
			if aware, ok := f.(DefaultPlagValuesAware); ok {
				if val, ok := aware.DefaultPlagValues(field.Name); ok {
					defaultVal = val.(string)
				}
			}
			p.StringVarP(pp.(*string), ss.ToSnake(field.Name), short, defaultVal, usage)
		case reflect.Bool:
			p.BoolVarP(pp.(*bool), ss.ToSnake(field.Name), short, false, usage)
		case reflect.Int:
			curDefault := 0
			switch defaultVal {
			case "runtime.GOMAXPROCS(0)":
				curDefault = runtime.GOMAXPROCS(0)
			}

			p.IntVarP(pp.(*int), ss.ToSnake(field.Name), short, curDefault, usage)
		case reflect.Slice:
			elemType := field.Type.Elem()
			switch elemType.Kind() {
			case reflect.String:
				curDefault := ss.Split(defaultVal, ",")
				if aware, ok := f.(DefaultPlagValuesAware); ok {
					if val, ok := aware.DefaultPlagValues(field.Name); ok {
						curDefault = val.([]string)
					}
				}
				p.StringSliceVarP(pp.(*[]string), ss.ToSnake(field.Name), short, curDefault, usage)
			}
		case reflect.Func:
			// ignore
		default:
			return fmt.Errorf(`unsupported type: %s, use kong:"-" to ignore`, field.Type)
		}
	}

	return nil
}

type DefaultPlagValuesAware interface {
	DefaultPlagValues(name string) (any, bool)
}
