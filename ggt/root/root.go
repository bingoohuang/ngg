package root

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/ver"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func CreateSubCmd(parent *cobra.Command, use, short string, obj interface {
	Run(*cobra.Command, []string) error
}) {
	c := &cobra.Command{
		Use:   use,
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			if err := obj.Run(cmd, args); err != nil {
				log.Printf("error occured: %v", err)
			}
		},
	}
	ss.PanicErr(InitFlags(obj, c.Flags(), c.PersistentFlags()))
	parent.AddCommand(c)
}

func AddCommand(c *cobra.Command, fc any) {
	if fc != nil && !c.DisableFlagParsing {
		ss.PanicErr(InitFlags(fc, c.Flags(), c.PersistentFlags()))
	}
	if runer, ok := fc.(interface {
		Run(cmd *cobra.Command, args []string) error
	}); ok {
		c.Run = func(cmd *cobra.Command, args []string) {
			if err := runer.Run(cmd, args); err != nil {
				log.Printf("error occured: %v", err)
			}
		}
	}
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

func InitFlags(f any, pf, persistent *pflag.FlagSet) error {
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
				InitFlags(squashField, pf, persistent)
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
		help := ""
		if v, _ := tags.Get("help"); v != nil {
			help = v.Raw
		}

		p := pf
		if v, _ := tags.Get("persistent"); v != nil {
			if persistentTag, _ := ss.ParseBool(v.Raw); persistentTag {
				p = persistent
			}
		}

		defaultVal := ""
		if v, _ := tags.Get("default"); v != nil {
			defaultVal = v.Raw
		}
		if defaultVal == "" {
			if t, _ := tags.Get("env"); t != nil && t.Raw != "-" {
				env := t.Raw
				if t.Raw == "auto" {
					env = ss.ToSnakeUpper(name)
				}
				defaultVal = os.Getenv(env)
				help = appendHelp(name, help, fmt.Sprintf("env: $%s.", env))
			}
		}
		var enumValues []string
		if v, _ := tags.Get("enum"); v != nil {
			enumValues = ss.Split(v.Raw, ",")
			help = appendHelp(name, help, fmt.Sprintf("allowed: %s.", v.Raw))
		}

		pp := structVal.Field(i).Addr().Interface()
		if pf, ok := pp.(pflag.Value); ok {
			if defaultVal != "" {
				pf.Set(defaultVal)
			}
			p.VarP(pf, name, short, help)
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
			p.DurationVarP(pp.(*time.Duration), name, short, curDefault, help)
			continue
		}

		switch field.Type.Kind() {
		case reflect.String:

			if aware, ok := f.(DefaultPlagValuesAware); ok {
				if val, ok := aware.DefaultPlagValues(field.Name); ok {
					defaultVal = val.(string)
				}
			}
			if len(enumValues) > 0 {
				p.VarP(NewEnum(enumValues, pp.(*string), defaultVal), name, short, help)
			} else {
				p.StringVarP(pp.(*string), name, short, defaultVal, help)
			}
		case reflect.Bool:
			curDefault := false
			if defaultVal != "" {
				curDefault, err = ss.ParseBool(defaultVal)
				if err != nil {
					return fmt.Errorf("parse  bool %s: %w", defaultVal, err)
				}
			}
			p.BoolVarP(pp.(*bool), name, short, curDefault, help)
		case reflect.Int:
			curDefault := 0
			if defaultVal != "" {
				switch defaultVal {
				case "runtime.GOMAXPROCS(0)":
					curDefault = runtime.GOMAXPROCS(0)
				default:
					curDefault, err = ss.Parse[int](defaultVal)
					if err != nil {
						log.Panicf("default %s is not int", defaultVal)
					}
				}
			}
			if len(enumValues) > 0 {
				p.VarP(NewEnumInt(enumValues, pp.(*int), curDefault), name, short, help)
			} else {
				p.IntVarP(pp.(*int), name, short, curDefault, help)
			}

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
				p.StringArrayVarP(pp.(*[]string), name, short, curDefault, help)
			}
		case reflect.Func:
			// ignore
		default:
			return fmt.Errorf(`unsupported type: %s, use kong:"-" to ignore`, field.Type)
		}
	}

	return nil
}

func appendHelp(name, help, s string) string {
	if help == "" {
		help = name
	}

	if !strings.HasSuffix(help, ".") {
		help += "."
	}

	return help + " " + s
}

type DefaultPlagValuesAware interface {
	DefaultPlagValues(name string) (any, bool)
}

type EnumType int

const (
	EnumString EnumType = iota
	EnumInt
)

type Enum struct {
	Allows    []string
	Value     *string
	ValueInt  *int
	ValueType EnumType
}

func NewEnum(allows []string, val *string, defaultVal string) *Enum {
	*val = defaultVal
	return &Enum{
		Allows:    allows,
		Value:     val,
		ValueType: EnumString,
	}
}

func NewEnumInt(allows []string, val *int, defaultVal int) *Enum {
	*val = defaultVal

	return &Enum{
		Allows:    allows,
		ValueInt:  val,
		ValueType: EnumInt,
	}
}

// String is used both by fmt.Print and by Cobra in help text
func (e Enum) String() string {
	switch e.ValueType {
	case EnumString:
		return *e.Value
	case EnumInt:
		return strconv.Itoa(*e.ValueInt)
	}
	return ""
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (e *Enum) Set(v string) error {
	for _, allow := range e.Allows {
		if strings.EqualFold(v, allow) {
			switch e.ValueType {
			case EnumString:
				*e.Value = v
				return nil
			case EnumInt:
				var err error
				*e.ValueInt, err = ss.Parse[int](v)
				return err
			}
		}
	}

	return errors.New(`must be one of ` + strings.Join(e.Allows, ","))
}

// Type is only used in help text
func (e *Enum) Type() string {
	return "enum"
}
