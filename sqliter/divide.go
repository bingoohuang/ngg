package sqliter

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/ss"
	"github.com/golang-module/carbon/v2"
)

// DividedBy 分库文件的时间分割模式
type DividedBy int

const (
	// DividedByMonth 按月
	DividedByMonth DividedBy = iota
	// DividedByWeek 按周
	DividedByWeek
	// DividedByDay 按天
	DividedByDay
)

// ErrUnknownDividedString 分割时间模式字符串无法识别
var ErrUnknownDividedString = errors.New("unknown divided string")

// ParseDivideString 解析分割时间模式字符串
func ParseDivideString(s string) (DividedBy, error) {
	switch strings.ToLower(s) {
	case "month":
		return DividedByMonth, nil
	case "week":
		return DividedByWeek, nil
	case "day":
		return DividedByDay, nil
	}

	return 0, fmt.Errorf("divied string %s is unkown: %w", s, ErrUnknownDividedString)
}

// DividedString 生成时间 t 的分割字符串
// 按月: month.yyyyMM
// 按周: week.yyyyWW
// 按天: day.yyyyMMdd
func (d DividedBy) DividedString(t time.Time) string {
	switch d {
	case DividedByMonth:
		return d.DividedPrefix() + t.Format("200601")
	case DividedByWeek:
		weekOfYear := carbon.CreateFromStdTime(t).WeekOfYear()
		return d.DividedPrefix() + t.Format("2006") + fmt.Sprintf("%02d", weekOfYear)
	case DividedByDay:
		return d.DividedPrefix() + t.Format("20060102")
	}

	log.Fatalf("unknown dividedBy %d", d)
	return ""
}

// DividedPrefix 时间分区的前缀
func (d DividedBy) DividedPrefix() string {
	switch d {
	case DividedByMonth:
		return "month."
	case DividedByWeek:
		return "week."
	case DividedByDay:
		return "day."
	}

	log.Fatalf("unknown dividedBy %d", d)
	return ""
}

// TableFilePath 返回表文件的完整前缀，例如: "testdata/metric.t.disk.month.202408.db"
func (q *Sqliter) TableFilePath(table, dividedBy string) string {
	return q.Prefix + q.TableFileBase(table, dividedBy)
}

// TableFileBase 返回表文件的基础文件名前缀，例如: "disk.month.202408.db"
func (q *Sqliter) TableFileBase(table, dividedBy string) string {
	return table + "." + dividedBy + ".db"
}

type TimeSpan struct {
	Value int
	Unit  TimeSpanUnit
}

var (
	unitReg        = regexp.MustCompile(`^\s*(?i)(\d+)\s*(month|months|m|week|weeks|w|days|day|d)\s*$`)
	ErrBadTimeSpan = errors.New("bad TimeSpan expr")
)

func ParseTimeSpan(s string) (TimeSpan, error) {
	subs := unitReg.FindStringSubmatch(s)
	if len(subs) == 0 {
		return TimeSpan{}, ErrBadTimeSpan
	}

	value, err := ss.Parse[int](subs[1])
	if err != nil {
		return TimeSpan{}, err
	}

	unit := UnitMonth
	switch strings.ToLower(subs[2][:0]) {
	case "m":
		unit = UnitMonth
	case "w":
		unit = UnitWeek
	case "d":
		unit = UnitDay
	}

	return TimeSpan{
		Value: value,
		Unit:  unit,
	}, nil
}

func (t *TimeSpan) String() string {
	switch t.Unit {
	case UnitMonth:
		return fmt.Sprintf("%d month", t.Value)
	case UnitWeek:
		return fmt.Sprintf("%d week", t.Value)
	case UnitDay:
		return fmt.Sprintf("%d day", t.Value)
	}
	return ""
}

// MinusBy 计算  c - t 的结果
func (t *TimeSpan) MinusBy(c carbon.Carbon) carbon.Carbon {
	switch t.Unit {
	case UnitMonth:
		return c.SubMonths(t.Value)
	case UnitWeek:
		return c.SubWeeks(t.Value)
	case UnitDay:
		return c.SubDays(t.Value)
	}

	return carbon.Carbon{}
}

type TimeSpanUnit int

const (
	UnitMonth TimeSpanUnit = iota
	UnitWeek
	UnitDay
)

func (u TimeSpanUnit) Of(value int) TimeSpan {
	return TimeSpan{
		Value: value,
		Unit:  u,
	}
}

// CutoffDays 根据时间 t, 以及保留天数 days, 计算切断时间点所在的划分时间值（如果等于当前时间划分值，则往前退一个时间划分）
func (d DividedBy) CutoffDays(t time.Time, timeSpan TimeSpan) string {
	ct := carbon.CreateFromStdTime(t)
	cutoffDay := timeSpan.MinusBy(ct)
	now := carbon.Now()

	cutoffDivided := d.DividedString(cutoffDay.StdTime())
	cutoffCurrent := d.DividedString(now.StdTime())
	if cutoffDivided == cutoffCurrent { // 调整为上个划分、
		switch d {
		case DividedByMonth:
			ct = ct.SubMonth()
		case DividedByWeek:
			ct = ct.SubWeek()
		case DividedByDay:
			ct = ct.SubDay()
		}
		cutoffDivided = d.DividedString(ct.StdTime())
	}
	return cutoffDivided
}
