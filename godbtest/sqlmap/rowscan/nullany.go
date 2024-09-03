package rowscan

import (
	"database/sql"
	"reflect"
	"time"
)

type NullAny struct {
	Value     any
	ValueType ValueType
	Valid     bool
}

// Scan implements the Scanner interface.
func (ns *NullAny) Scan(value any) error {
	ns.Valid = value != nil
	if !ns.Valid {
		return nil
	}

	switch ns.ValueType {
	case ValueTypeBool:
		var v0 sql.NullBool
		err := v0.Scan(value)
		ns.Value = v0.Bool
		return err
	case ValueTypeInt64:
		var v0 sql.NullInt64
		err := v0.Scan(value)
		ns.Value = v0.Int64
		return err
	case ValueTypeFloat64:
		var v1 sql.NullFloat64
		err := v1.Scan(value)
		ns.Value = v1.Float64
		return err
	case ValueTypeString:
		var v2 sql.NullString
		err := v2.Scan(value)
		ns.Value = v2.String
		return err
	case ValueTypeBytes:
		if _, ok := value.([]byte); ok {
			ns.Value = value
			return nil
		}
		fallthrough
	default:
		switch nv := value.(type) {
		case int8, int16, int32, int, int64, uint8, uint16, uint32, uint, uint64:
			ns.ValueType = ValueTypeInt64
			ns.Value = nv
		case float32, float64:
			ns.ValueType = ValueTypeFloat64
			ns.Value = nv
		case string:
			ns.ValueType = ValueTypeString
			ns.Value = nv
		case []byte:
			if len(nv) < 1024 {
				if ns.ValueType == ValueTypeOther {
					ns.ValueType = ValueTypeString
					ns.Value = string(nv)
				}
			} else {
				ns.ValueType = ValueTypeBytes
				ns.Value = nv
			}
		default:
			ns.convertAlias(value)
		}

		return nil
	}
}

func (ns *NullAny) convertAlias(value any) {
	rv := reflect.ValueOf(value)

	for _, alias := range aliasConverters {
		if rv.CanConvert(alias.Type) {
			alias.Converter(alias.Type, rv, ns)
			return
		}
	}

	ns.Value = value
}

type AliasConvert struct {
	Type      reflect.Type
	Converter func(typ reflect.Type, value reflect.Value, ns *NullAny)
}

var aliasConverters = []AliasConvert{
	{
		Type: reflect.TypeOf((*time.Time)(nil)).Elem(),
		Converter: func(typ reflect.Type, rv reflect.Value, ns *NullAny) {
			t := rv.Convert(typ).Interface().(time.Time)

			// refer: https://github.com/hexon/mysqltsv/blob/main/mysqltsv.go
			hour, minute, sec := t.Clock()

			switch nsec := t.Nanosecond(); {
			case hour == 0 && minute == 0 && sec == 0 && nsec == 0:
				ns.Value = t.Format("2006-01-02")
			case nsec == 0:
				ns.Value = t.Format("2006-01-02 15:04:05")
			default:
				ns.Value = t.Format("2006-01-02 15:04:05.999999999")
			}
			ns.ValueType = ValueTypeString
		},
	},

	{
		Type: reflect.TypeOf(""),
		Converter: func(typ reflect.Type, rv reflect.Value, ns *NullAny) {
			ns.ValueType = ValueTypeString
			ns.Value = rv.Convert(typ).Interface().(string)
		},
	},
}
