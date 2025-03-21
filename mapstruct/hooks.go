package mapstruct

import (
	"encoding"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// typedDecodeHook takes a raw HookFunc (an any) and turns
// it into the proper HookFunc type, such as HookFuncType.
func typedDecodeHook(h HookFunc) HookFunc {
	// Create variables here so we can reference them with the reflect pkg
	var f1 HookFuncType
	var f2 HookFuncKind
	var f3 HookFuncValue

	// Fill in the variables into this interface and the rest is done
	// automatically using the reflect package.
	potential := []any{f1, f2, f3}

	v := reflect.ValueOf(h)
	vt := v.Type()
	for _, raw := range potential {
		pt := reflect.ValueOf(raw).Type()
		if vt.ConvertibleTo(pt) {
			return v.Convert(pt).Interface()
		}
	}

	return nil
}

// DecodeHookExec executes the given decode hook. This should be used
// since it'll naturally degrade to the older backwards compatible HookFunc
// that took reflect.Kind instead of reflect.Type.
func DecodeHookExec(raw HookFunc, from, to reflect.Value) (any, error) {
	switch f := typedDecodeHook(raw).(type) {
	case HookFuncType:
		return f(from.Type(), to.Type(), from.Interface())
	case HookFuncKind:
		return f(from.Kind(), to.Kind(), from.Interface())
	case HookFuncValue:
		return f(from, to)
	default:
		return nil, errors.New("invalid decode hook signature")
	}
}

// ComposeDecodeHookFunc creates a single HookFunc that
// automatically composes multiple DecodeHookFuncs.
//
// The composed funcs are called in order, with the result of the
// previous transformation.
func ComposeDecodeHookFunc(fs ...HookFunc) HookFunc {
	return func(f reflect.Value, t reflect.Value) (any, error) {
		var err error
		var data any
		newFrom := f
		for _, f1 := range fs {
			data, err = DecodeHookExec(f1, newFrom, t)
			if err != nil {
				return nil, err
			}
			newFrom = reflect.ValueOf(data)
		}

		return data, nil
	}
}

// StringToSliceHookFunc returns a HookFunc that converts
// string to []string by splitting on the given sep.
func StringToSliceHookFunc(sep string) HookFunc {
	return func(
		f reflect.Kind,
		t reflect.Kind,
		data any,
	) (any, error) {
		if f != reflect.String || t != reflect.Slice {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return []string{}, nil
		}

		return strings.Split(raw, sep), nil
	}
}

// StringToTimeDurationHookFunc returns a HookFunc that converts
// strings to time.Duration.
func StringToTimeDurationHookFunc() HookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(time.Duration(5)) {
			return data, nil
		}

		// Convert it by parsing
		return time.ParseDuration(data.(string))
	}
}

// StringToIPHookFunc returns a HookFunc that converts
// strings to net.IP
func StringToIPHookFunc() HookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(net.IP{}) {
			return data, nil
		}

		// Convert it by parsing
		ip := net.ParseIP(data.(string))
		if ip == nil {
			return net.IP{}, fmt.Errorf("failed parsing ip %v", data)
		}

		return ip, nil
	}
}

// StringToIPNetHookFunc returns a HookFunc that converts
// strings to net.IPNet
func StringToIPNetHookFunc() HookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(net.IPNet{}) {
			return data, nil
		}

		// Convert it by parsing
		_, n, err := net.ParseCIDR(data.(string))
		return n, err
	}
}

// StringToTimeHookFunc returns a HookFunc that converts
// strings to time.Time.
func StringToTimeHookFunc(layout string) HookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		// Convert it by parsing
		return time.Parse(layout, data.(string))
	}
}

// WeaklyTypedHook is a HookFunc which adds support for weak typing to
// the decoder.
//
// Note that this is significantly different from the WeakType option
// of the Config.
func WeaklyTypedHook(
	f reflect.Kind,
	t reflect.Kind,
	data any,
) (any, error) {
	dataVal := reflect.ValueOf(data)
	switch t {
	case reflect.String:
		switch f {
		case reflect.Bool:
			if dataVal.Bool() {
				return "1", nil
			}
			return "0", nil
		case reflect.Float32:
			return strconv.FormatFloat(dataVal.Float(), 'f', -1, 64), nil
		case reflect.Int:
			return strconv.FormatInt(dataVal.Int(), 10), nil
		case reflect.Slice:
			dataType := dataVal.Type()
			elemKind := dataType.Elem().Kind()
			if elemKind == reflect.Uint8 {
				return string(dataVal.Interface().([]uint8)), nil
			}
		case reflect.Uint:
			return strconv.FormatUint(dataVal.Uint(), 10), nil
		}
	}

	return data, nil
}

func RecursiveStructToMapHookFunc() HookFunc {
	return func(f reflect.Value, t reflect.Value) (any, error) {
		if f.Kind() != reflect.Struct {
			return f.Interface(), nil
		}

		var i any = struct{}{}
		if t.Type() != reflect.TypeOf(&i).Elem() {
			return f.Interface(), nil
		}

		m := make(map[string]any)
		t.Set(reflect.ValueOf(m))

		return f.Interface(), nil
	}
}

// TextUnmarshallerHookFunc returns a HookFunc that applies
// strings to the UnmarshalText function, when the target type
// implements the encoding.TextUnmarshaler interface
func TextUnmarshallerHookFunc() HookFuncType {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any,
	) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		result := reflect.New(t).Interface()
		unmarshaller, ok := result.(encoding.TextUnmarshaler)
		if !ok {
			return data, nil
		}
		if err := unmarshaller.UnmarshalText([]byte(data.(string))); err != nil {
			return nil, err
		}
		return result, nil
	}
}
