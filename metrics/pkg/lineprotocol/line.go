package lineprotocol

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Build format inputs to line protocol
// https://docs.influxdata.com/influxdb/v1.7/write_protocols/line_protocol_tutorial/
func Build(name string, tags map[string]string, fields map[string]interface{}, t time.Time) (string, error) {
	if len(fields) == 0 {
		return "", errors.New("fields are empty")
	}

	tagstr := ""

	for k, v := range tags {
		if k != "" && v != "" {
			tagstr += fmt.Sprintf(",%s=%s", escapeSpecialChars(k), escapeSpecialChars(v))
		}
	}

	out := ""
	// serialize fields
	for k, v := range fields {
		s, err := toInfluxRepr(v)
		if err != nil {
			return "", fmt.Errorf("toInfluxRepr error %w", err)
		}
		out += fmt.Sprintf(",%s=%s", escapeSpecialChars(k), s)
	}

	if out != "" {
		out = out[1:]
	}

	// construct line protocol string
	return fmt.Sprintf("%s%s %s %d", name, tagstr, out, uint64(t.UnixNano())), nil
}

func escapeSpecialChars(in string) string {
	str := strings.Replace(in, ",", `\,`, -1)
	str = strings.Replace(str, "=", `\=`, -1)
	str = strings.Replace(str, " ", `\ `, -1)

	return str
}

// toInfluxRepr 将val转换为Influx表示形式
func toInfluxRepr(val interface{}) (string, error) {
	switch v := val.(type) {
	case string:
		return strToInfluxRepr(v)
	case []byte:
		return strToInfluxRepr(string(v))
	case int32, int64, int16, int8, int, uint32, uint64, uint16, uint8, uint:
		return fmt.Sprintf("%d", v), nil
	case float64, float32:
		return fmt.Sprintf("%g", v), nil
	case bool:
		return fmt.Sprintf("%t", v), nil
	case time.Time:
		return fmt.Sprintf("%d", uint64(v.UnixNano())), nil
	case time.Duration:
		return fmt.Sprintf("%d", uint64(v.Nanoseconds())), nil
	default:
	}

	if s, ok := val.(fmt.Stringer); ok {
		return strToInfluxRepr(s.String())
	}

	return "", fmt.Errorf("%+v: unsupported type for Influx Line Protocol", val)
}

func strToInfluxRepr(v string) (string, error) {
	if len(v) > 64000 { // nolint gomnd
		return "", fmt.Errorf("string too long (%d characters, max. 64K)", len(v))
	}

	return fmt.Sprintf("%q", v), nil
}
