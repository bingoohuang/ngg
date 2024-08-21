package influx

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Metric 指标接口
type Metric interface {
	// Name 指标名称
	Name() string
	// Tags 指标标签（索引字段）
	Tags() map[string]string
	// Fields 指标字段（普通字段）
	Fields() map[string]any
	// Time 指标时间
	Time() time.Time
}

type Point struct {
	MetricName   string            `json:"name"`
	MetricTags   map[string]string `json:"tags"`
	MetricFields map[string]any    `json:"fields"`
	Timestamp    int64             `json:"timestamp"`
	MetricTime   time.Time
}

var _ Metric = (*Point)(nil)

func (m *Point) Name() string            { return m.MetricName }
func (m *Point) Tags() map[string]string { return m.MetricTags }
func (m *Point) Fields() map[string]any  { return m.MetricFields }
func (m *Point) Time() time.Time         { return m.MetricTime }

func NewPoint(name string, tags map[string]string, fields map[string]any, unix time.Time) Metric {
	return &Point{
		MetricName:   name,
		MetricTags:   tags,
		MetricFields: fields,
		Timestamp:    unix.Unix(),
		MetricTime:   unix,
	}
}

// ParseLineProtocol 解析InfluxDB的行协议字符串
func ParseLineProtocol(lp string) (*Point, error) {
	parts := strings.SplitN(lp, " ", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid line protocol format")
	}

	measurement, tags, err := parseTags(parts[0])
	if err != nil {
		return nil, err
	}
	fields, err := parseFields(parts[1])
	if err != nil {
		return nil, err
	}

	// 解析时间戳
	timestamp, err := parseTimestamp(parts[2])
	if err != nil {
		return nil, err
	}

	return &Point{
		MetricName:   measurement,
		MetricTags:   tags,
		MetricFields: fields,
		Timestamp:    timestamp,
		MetricTime:   time.Unix(timestamp/1000000000, 0),
	}, nil
}

// parseTags 解析标签字符串
func parseTags(tagStr string) (string, map[string]string, error) {
	tags := make(map[string]string)
	parts := strings.Split(tagStr, ",")
	for _, tag := range parts[1:] {
		parts := strings.SplitN(tag, "=", 2)
		if len(parts) != 2 {
			return "", nil, fmt.Errorf("invalid tag format: %s", tag)
		}
		tags[parts[0]] = parts[1]
	}
	return parts[0], tags, nil
}

// parseFields 解析字段字符串
func parseFields(fieldStr string) (map[string]any, error) {
	fields := make(map[string]any)
	for _, field := range strings.Split(fieldStr, ",") {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid field format: %s", field)
		}
		value, err := parseFieldValue(parts[1])
		if err != nil {
			return nil, err
		}
		fields[parts[0]] = value
	}
	return fields, nil
}

// parseFieldValue 解析字段值
func parseFieldValue(valueStr string) (any, error) {
	// 这里可以添加更多的类型解析，例如布尔值、浮点数等
	switch valueStr {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		if strings.HasSuffix(valueStr, "i") {
			// 如果值以 'i' 结尾，表示它是一个整数
			valueStr = valueStr[:len(valueStr)-1] // 去除 'i'
			value, err := strconv.ParseInt(valueStr, 10, 64)
			return value, err
		}

		// 默认尝试解析为浮点数
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid field value: %s", valueStr)
		}
		return value, nil
	}
}

// parseTimestamp 解析时间戳
func parseTimestamp(timestampStr string) (int64, error) {
	// InfluxDB的时间戳通常是Unix时间戳，单位为纳秒
	return strconv.ParseInt(timestampStr, 10, 64)
}
