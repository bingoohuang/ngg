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

func AddCommand(c *cobra.Command, fc any) {
	ss.PanicErr(initFlags(fc, c.Flags()))
	cmd.AddCommand(c)
}

var cmd = func() *cobra.Command {
	r := &cobra.Command{
		Use:   "ggt",
		Short: "golang toolset",
	}

	r.Version = "version"
	r.SetVersionTemplate(ver.Version() + "\n")

	return r
}()

func Run() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

func initFlags(f any, p *pflag.FlagSet) error {
	ptrVal := reflect.ValueOf(f)
	structVal := ptrVal.Elem()
	structType := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}

		tags, err := ss.ParseStructTags(string(field.Tag))
		if err != nil {
			return err
		}

		if kong, _ := tags.Get("kong"); kong != nil && kong.Raw == "-" {
			continue
		}

		if field.Anonymous {
			if tags.GetTag("squash") == "true" {
				squashField := structVal.Field(i).Addr().Interface()
				initFlags(squashField, p)
			}
			continue
		}

		name := ss.ToSnake(field.Name)
		if v, _ := tags.Get("flag"); v != nil {
			name = v.Raw
		}

		short := ""
		if v, _ := tags.Get("short"); v != nil {
			short = v.Raw
		}
		usage := ""
		if v, _ := tags.Get("help"); v != nil {
			usage = v.Raw
		}
		defaultVal := ""
		if v, _ := tags.Get("default"); v != nil {
			defaultVal = v.Raw
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
			p.VarP(pf, name, short, usage)
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
			p.DurationVarP(pp.(*time.Duration), name, short, curDefault, usage)
			continue
		}

		switch field.Type.Kind() {
		case reflect.String:
			if aware, ok := f.(DefaultPlagValuesAware); ok {
				if val, ok := aware.DefaultPlagValues(field.Name); ok {
					defaultVal = val.(string)
				}
			}
			p.StringVarP(pp.(*string), name, short, defaultVal, usage)
		case reflect.Bool:
			p.BoolVarP(pp.(*bool), name, short, false, usage)
		case reflect.Int:
			curDefault := 0
			if defaultVal != "" {
				switch defaultVal {
				case "runtime.GOMAXPROCS(0)":
					curDefault = runtime.GOMAXPROCS(0)
				default:
					intVal, err := ss.Parse[int](defaultVal)
					if err != nil {
						return fmt.Errorf("parse int %s: %w", defaultVal, err)
					}
					curDefault = intVal
				}
			}

			p.IntVarP(pp.(*int), name, short, curDefault, usage)
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
				p.StringArrayVarP(pp.(*[]string), name, short, curDefault, usage)
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
