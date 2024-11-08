package ss

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
)

// StructEnv 通过对应的环境变量的值，设置结构体字段的值.
// 结构体字段支持以下类型或者该类型的指针:
// string, bool,
// uint64, uint32, uint16, uint8, uint,
// int64, int32, int16, int8, int,
// float32, float64
// time.Duration
// 可以通过在 Tag 中使用 env:"-" 忽略某个字段。或者 env:"ENV_NAME" 来指定环境变量名。
// 如果所设置的环境变量值类型不符合，将会返回错误
// 参数 v 必须是结构体变量的指针。
// 示例:
//
//	type Config struct {
//		Port int
//	}
//	func main() {
//		var cfg Config
//		StructEnv(&cfg, &StructEnvOptions{Prefix: "MYAPP_", PanicOnError: true})
//		// 环境变量 export MYAPP_PORT = 10000
//		fmt.Println(cfg.Port) // 会打印 10000
//	}
func StructEnv(v any, envOptions ...*StructEnvOptions) error {
	if len(envOptions) == 0 {
		envOptions = []*StructEnvOptions{{}}
	}

	opt := envOptions[0]
	err := structEnv(v, opt)
	if err != nil && opt.PanicOnError {
		panic(err)
	}
	return err
}

func structEnv(v any, opt *StructEnvOptions) error {

	reflectVal := reflect.ValueOf(v)
	if reflectVal.Kind() != reflect.Ptr {
		return fmt.Errorf("v must be pointer")
	}

	valueElem := reflectVal.Elem()
	for i := 0; i < valueElem.NumField(); i++ {
		structField := valueElem.Type().Field(i)
		if !structField.IsExported() {
			continue
		}

		envName := structField.Tag.Get("env")
		if envName == "-" {
			continue
		}
		if envName == "" {
			envName = ToSnakeUpper(structField.Name)
		}

		fieldType := structField.Type
		kind := fieldType.Kind()
		isPtr := kind == reflect.Ptr
		if isPtr {
			fieldType = fieldType.Elem()
			kind = fieldType.Kind()
		}

		field := valueElem.Field(i)
		envName = opt.Prefix + envName
		envValue := os.Getenv(envName)
		if envValue == "" {
			if fieldType == timeDurationType {
				if Pick1(ParseBool(os.Getenv("STRUCT_ENV_VERBOSE"))) {
					log.Printf("Can set env %s (%s), current: %v", envName, fieldType, field.Interface())
				}
			} else {
				switch kind {
				case reflect.String, reflect.Bool,
					reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint,
					reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int,
					reflect.Float32, reflect.Float64:
					if Pick1(ParseBool(os.Getenv("STRUCT_ENV_VERBOSE"))) {
						log.Printf("Can set env %s (%s), current: %v", envName, kind, field.Interface())
					}
				default:
					// ignore other types
				}
			}

			continue
		}

		log.Printf("Will set env %s set to %s", envName, envValue)

		if fieldType == timeDurationType {
			d, _, err := ParseDur(envValue)
			if err != nil {
				return fmt.Errorf("parse %s = %s as Duration error: %v", envName, envValue, err)
			}

			if isPtr {
				field.Set(reflect.ValueOf(&d))
			} else {
				field.Set(reflect.ValueOf(d))
			}
			continue
		}

		switch kind {
		case reflect.String:
			if isPtr {
				field.Set(reflect.ValueOf(&envValue))
			} else {
				field.SetString(envValue)
			}
		case reflect.Bool:
			ev, err := ParseBool(envValue)
			if err != nil {
				return fmt.Errorf("parse %s = %s as Bool error: %v", envName, envValue, err)
			}
			if isPtr {
				field.Set(reflect.ValueOf(&ev))
			} else {
				field.SetBool(ev)
			}
		case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
			ev, err := strconv.ParseUint(envValue, 10, 64)
			if err != nil {
				return fmt.Errorf("parse %s = %s as Uint64 error: %v", envName, envValue, err)
			}
			if isPtr {
				field.Set(reflect.ValueOf(&ev))
			} else {
				field.SetUint(ev)
			}
		case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
			ev, err := strconv.ParseInt(envValue, 10, 64)
			if err != nil {
				return fmt.Errorf("parse %s = %s as Int64 error: %v", envName, envValue, err)
			}
			if isPtr {
				field.Set(reflect.ValueOf(&ev))
			} else {
				field.SetInt(ev)
			}

		case reflect.Float32, reflect.Float64:
			ev, err := strconv.ParseFloat(envValue, 64)
			if err != nil {
				return fmt.Errorf("parse %s = %s as Float64 error: %v", envName, envValue, err)
			}
			if isPtr {
				field.Set(reflect.ValueOf(&ev))
			} else {
				field.SetFloat(ev)
			}
		default:
			// ignore other types
		}
	}

	return nil
}

// StructEnvOptions is the options for StructEnv.
type StructEnvOptions struct {
	// Prefix is the prefix of environment variable name.
	Prefix string
	// PanicOnError will panic if there is an error.
	PanicOnError bool
}
