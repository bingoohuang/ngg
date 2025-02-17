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

func RunCmd(parent *cobra.Command, use, short string, obj interface {
	Run(*cobra.Command, []string) error
}) {
	c := CreateCmd(parent, use, short, obj)
	if err := c.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

type CmdLongHelper interface {
	LongHelp() string
}

func CreateCmd(parent *cobra.Command, use, short string, obj interface {
	Run(*cobra.Command, []string) error
}) *cobra.Command {
	longHelp := ""
	if l, ok := obj.(CmdLongHelper); ok {
		longHelp = l.LongHelp()
	}

	c := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  longHelp,
		Run: func(cmd *cobra.Command, args []string) {
			if err := obj.Run(cmd, args); err != nil {
				log.Printf("error occured: %v", err)
			}
		},
	}

	ss.PanicErr(InitFlags(obj, c.Flags(), c.PersistentFlags()))

	if parent != nil {
		parent.AddCommand(c)
	}

	return c
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
			if ss.Pick1(ss.ParseBool(tags.GetTag("squash"))) {
				squashField := structVal.Field(i).Addr().Interface()
				InitFlags(squashField, pf, persistent)
			}
			continue
		}

		name := ss.ToKebab(field.Name)
		if v, _ := tags.Get("flag"); v != nil {
			name = v.Raw
		}
		short := tags.GetTag("short")
		help := tags.GetTag("help")

		p := pf
		if v, _ := tags.Get("persistent"); v != nil {
			if persistentTag, _ := ss.ParseBool(v.Raw); persistentTag {
				p = persistent
			}
		}

		var defaultVal string
		if t, _ := tags.Get("env"); t != nil && t.Raw != "-" {
			env := t.Raw
			if t.Raw == "auto" {
				env = ss.ToSnakeUpper(name)
			}
			for _, envName := range ss.Split(env, ",") {
				if envVal := os.Getenv(envName); envVal != "" {
					defaultVal = envVal
					help = appendHelp(name, help, fmt.Sprintf("env: %s.", env))
					break
				}
			}
		}
		if defaultVal == "" {
			defaultVal = tags.GetTag("default")
		}

		var enumValues []string
		if v, _ := tags.Get("enum"); v != nil {
			enumValues = SplitSeps(v.Raw, ",")
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

		var awareValue any
		if aware, ok := f.(DefaultPlagValuesAware); ok {
			if val, ok := aware.DefaultPlagValues(field.Name); ok {
				awareValue = val
			}
		}

		switch field.Type.Kind() {
		case reflect.String:
			if awareValue != nil {
				defaultVal = awareValue.(string)
			}
			if len(enumValues) > 0 {
				p.VarP(NewEnum(enumValues, pp.(*string), defaultVal), name, short, help)
			} else {
				p.StringVarP(pp.(*string), name, short, defaultVal, help)
			}
		case reflect.Bool:
			if v, _ := tags.Get("version"); v != nil {
				flag := p.VarPF(NewVersionValue(), name, short, help)
				flag.NoOptDefVal = "true"
				continue
			}

			curDefault := false
			if awareValue != nil {
				curDefault = awareValue.(bool)
			} else if defaultVal != "" {
				curDefault, err = ss.ParseBool(defaultVal)
				if err != nil {
					return fmt.Errorf("parse  bool %s: %w", defaultVal, err)
				}
			}
			p.BoolVarP(pp.(*bool), name, short, curDefault, help)
		case reflect.Int:
			curDefault := 0
			if awareValue != nil {
				curDefault = awareValue.(int)
			} else if defaultVal != "" {
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

			if ss.Pick1(ss.ParseBool(tags.GetTag("count"))) {
				p.CountVarP(pp.(*int), name, short, help)
			} else {
				if len(enumValues) > 0 {
					p.VarP(NewEnumInt(enumValues, pp.(*int), curDefault), name, short, help)
				} else {
					p.IntVarP(pp.(*int), name, short, curDefault, help)
				}
			}
		case reflect.Int32:
			curDefault := int32(0)
			if awareValue != nil {
				curDefault = awareValue.(int32)
			} else if defaultVal != "" {
				curDefault, err = ss.Parse[int32](defaultVal)
				if err != nil {
					log.Panicf("default %s is not int", defaultVal)
				}
			}

			p.Int32VarP(pp.(*int32), name, short, curDefault, help)
		case reflect.Int64:
			curDefault := int64(0)
			if awareValue != nil {
				curDefault = awareValue.(int64)
			} else if defaultVal != "" {
				curDefault, err = ss.Parse[int64](defaultVal)
				if err != nil {
					log.Panicf("default %s is not int", defaultVal)
				}
			}

			p.Int64VarP(pp.(*int64), name, short, curDefault, help)
		case reflect.Float32:
			curDefault := float32(0)
			if awareValue != nil {
				curDefault = awareValue.(float32)
			} else if defaultVal != "" {
				curDefault, err = ss.Parse[float32](defaultVal)
				if err != nil {
					log.Panicf("default %s is not int", defaultVal)
				}
			}

			p.Float32VarP(pp.(*float32), name, short, curDefault, help)
		case reflect.Float64:
			curDefault := float64(0)
			if awareValue != nil {
				curDefault = awareValue.(float64)
			} else if defaultVal != "" {
				curDefault, err = ss.Parse[float64](defaultVal)
				if err != nil {
					log.Panicf("default %s is not int", defaultVal)
				}
			}

			p.Float64VarP(pp.(*float64), name, short, curDefault, help)
		case reflect.Slice:
			elemType := field.Type.Elem()
			switch elemType.Kind() {
			case reflect.String:
				curDefault := ss.Split(defaultVal, ",")
				if awareValue != nil {
					curDefault = awareValue.([]string)
				}
				p.StringArrayVarP(pp.(*[]string), name, short, curDefault, help)
			}
		case reflect.Func, reflect.Map, reflect.Chan, reflect.Interface, reflect.UnsafePointer:
			// ignore
		default:
			return fmt.Errorf(`unsupported type: %s %s, use kong:"-" to ignore`, field.Name, field.Type)
		}
	}

	return nil
}

func SplitSeps(s string, seps string) []string {
	var v []string
	for _, f := range strings.Split(s, seps) {
		f = strings.TrimSpace(f)
		v = append(v, f)
	}

	return v
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

type versionValue struct{}

func NewVersionValue() *versionValue {
	return &versionValue{}
}

func (b *versionValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if v {
		fmt.Printf("%s\n", ver.Version())
		os.Exit(0)
	}
	return err
}

func (b *versionValue) Type() string     { return "bool" }
func (b *versionValue) String() string   { return strconv.FormatBool(false) }
func (b *versionValue) IsBoolFlag() bool { return true }
