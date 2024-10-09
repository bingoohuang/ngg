package metric

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/metrics/pkg/ks"
	"github.com/bingoohuang/ngg/metrics/pkg/lineprotocol"
)

// LogType means the logMetrics type.
type LogType string

const (
	// KeyRT RT 日志类型.
	KeyRT LogType = "RT"
	// KeyQPS QPS 日志类型.
	KeyQPS = "QPS"
	// KeySuccessRate SuccessRate 日志类型.
	KeySuccessRate = "SUCCESS_RATE"
	// KeyFailRate FailRate 日志类型.
	KeyFailRate = "FAIL_RATE"
	// KeyHitRate HitRate 日志类型.
	KeyHitRate = "HIT_RATE"
	// KeyCUR CUR 日志类型.
	KeyCUR = "CUR"

	// HB 特殊处理，每?s记录一次.
	HB = "HB"
)

// isSimple 是否简单的值，值与值之间，不需要有累计等关系.
func (lt LogType) isSimple() bool { return lt == KeyCUR }

// isPercent 是否是百分比类型.
func (lt LogType) isPercent() bool {
	switch lt {
	case KeySuccessRate, KeyFailRate, KeyHitRate:
		return true
	default:
		return false
	}
}

const TimeLayout = "20060102150405000"

// Line represents a metric rotate line structure in rotate file.
type Line struct {
	Ks *ks.Ks `json:"-"`

	K1  string `json:"k1,omitempty"`
	K2  string `json:"k2,omitempty"`
	K3  string `json:"k3,omitempty"`
	K4  string `json:"k4,omitempty"`
	K5  string `json:"k5,omitempty"`
	K6  string `json:"k6,omitempty"`
	K7  string `json:"k7,omitempty"`
	K8  string `json:"k8,omitempty"`
	K9  string `json:"k9,omitempty"`
	K10 string `json:"k10,omitempty"`
	K11 string `json:"k11,omitempty"`
	K12 string `json:"k12,omitempty"`
	K13 string `json:"k13,omitempty"`
	K14 string `json:"k14,omitempty"`
	K15 string `json:"k15,omitempty"`
	K16 string `json:"k16,omitempty"`
	K17 string `json:"k17,omitempty"`
	K18 string `json:"k18,omitempty"`
	K19 string `json:"k19,omitempty"`
	K20 string `json:"k20,omitempty"`

	Key      string  `json:"key"` // {{k1}}#{{k2}}#{{k3}}
	LogType  LogType `json:"logtype"`
	Time     string  `json:"time"` // yyyyMMddHHmmssSSS
	Hostname string  `json:"hostname"`

	Keys []string `json:"-"`
	Min  float64  `json:"min"` // 每次采集区间（METRICS_INTERVAL）中 v1  最小/大值
	Max  float64  `json:"max"` // 只对 RT 生效

	V1 float64 `json:"v1"` // 小数
	V2 float64 `json:"v2"` // 只有比率类型的时候，才用到v2
	V3 float64 `json:"v3"` // RT 当 [300-400) ms 时 v3 = 1
	V4 float64 `json:"v4"` // RT 当 [400-500) ms 时 v4 = 1
	V5 float64 `json:"v5"` // RT 当 [500-600) ms 时 v5 = 1
	V6 float64 `json:"v6"` // RT 当 [600-700) ms 时 v6 = 1
	V7 float64 `json:"v7"` // RT 当 [700-800) ms 时 v7 = 1
	V8 float64 `json:"v8"` // RT 当 [800-900) ms 时 v8 = 1
	V9 float64 `json:"v9"` // RT 当 [900-∞) ms 时 v9 = 1

	N int `json:"n"` // 上次记录完日志后，累积的写次数
}

// ToLineProtocol print l to a influxdb v1 line protocol format.
func (l *Line) ToLineProtocol() (string, error) {
	t, err := time.Parse(TimeLayout, l.Time)
	if err != nil {
		return "", err
	}

	fields := map[string]interface{}{
		"v1": l.V1, "v2": l.V2,
		"v3": l.V3, "v4": l.V4, "v5": l.V5, "v6": l.V6, "v7": l.V7, "v8": l.V8, "v9": l.V9,
	}

	if l.LogType == KeyRT {
		fields["min"] = l.Min
		fields["max"] = l.Max
	}

	if l.Ks != nil {
		for i := 3; i <= len(l.Ks.Keys); i++ {
			if l.Ks.Keys[i-1] != "" {
				fields[fmt.Sprintf("k%d", i)] = l.Ks.Keys[i-1]
			}
		}
	}

	return lineprotocol.Build(string(l.LogType),
		map[string]string{"key": l.Key, "hostname": l.Hostname},
		fields,
		t)
}

func (l *Line) hasExtraKeys() bool {
	if l.Ks != nil {
		for i := 3; i <= len(l.Ks.Keys); i++ {
			if l.Ks.Keys[i-1] != "" {
				return true
			}
		}
	}

	return false
}

func (l *Line) fulfilKeys() {
	if len(l.Keys) >= 3 {
		l.K1 = l.Keys[0]
		l.K2 = l.Keys[1]
		l.K3 = l.Keys[2]
	} else if len(l.Keys) >= 2 {
		l.K1 = l.Keys[0]
		l.K2 = l.Keys[1]
	} else if len(l.Keys) >= 1 {
		l.K1 = l.Keys[0]
	}

	lv := reflect.ValueOf(l).Elem()

	if l.Ks != nil {
		for i := 4; i <= 20; i++ {
			k := fmt.Sprintf("K%d", i)
			lv.FieldByName(k).Set(reflect.ValueOf(l.Ks.Keys[i-1]))
		}
	}
}

// AsyncPut new a metric line.
func (r *Runner) AsyncPut(keys Key, logType LogType, v1, v2 float64, vx ...float64) {
	fv := func(idx int) float64 {
		if len(vx) > idx-3 {
			return vx[idx-3]
		}

		return 0
	}

	line := &Line{
		Keys:    keys.Keys,
		Key:     strings.Join(keys.Keys, "#"),
		LogType: logType,
		V1:      v1,
		V2:      v2,
		V3:      fv(3),
		V4:      fv(4),
		V5:      fv(5),
		V6:      fv(6),
		V7:      fv(7),
		V8:      fv(8),
		V9:      fv(9),
		Ks:      keys.ks,
	}
	if r.autoDrop {
		select {
		case r.C <- line:
		default: // bypass, async.
		}
	} else {
		r.C <- line
	}
}
