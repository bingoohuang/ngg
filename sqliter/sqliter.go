package sqliter

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// New 创建 *sqliter 对象
func New(fns ...ConfigFn) (*Sqliter, error) {
	config, err := createConfig(fns)
	if err != nil {
		return nil, err
	}

	plus := &Sqliter{
		Config:   config,
		readDbs:  make(map[string]*readTable),
		writeDbs: make(map[string]*writeTable),
	}
	go plus.recycleLoop()

	return plus, nil
}

type Config struct {
	// DriverName 驱动名称, 例如 sqlite3
	DriverName string
	// Prefix 设定库文件的前缀(包括完整路径）
	Prefix string
	// WriteDsnOptions 连接字符串选项，比如 _journal=WAL
	WriteDsnOptions string
	// ReadDsnOptions 连接字符串选项，比如 _txlock=immediate
	ReadDsnOptions string
	// BadTableSubs 非法表名子串
	BadTableSubs []string

	// AllowDBErrors 允许的DB错误字眼，否则被认为库文件损坏
	AllowDBErrors []string
	// MaxIdle 最大数据库读写空闲时间，超期关闭数据库
	MaxIdle time.Duration

	// BatchInsertInterval 批量插入时间间隔
	BatchInsertInterval time.Duration
	// BatchInsertSize 批量插入大小
	BatchInsertSize int

	// SeqKeysDBName keys 字符串转换为枚举数字的 boltdb 库名字（可以包括路径），默认 keyseq.bolt
	// 为 sqlite 的 tag 字符串值生成唯一的对应序号（减少sqlite数据库文件大小而优化设计）
	SeqKeysDBName string

	// SeqKeysDB 是 SeqKeysDB 对应的对象
	SeqKeysDB *BoltSeq
	// Debug 是否开启 Debug 模式，打印 SQL 等
	Debug bool

	// AsTags 用来转换普通字段为索引字段的判断器
	AsTags Matcher

	// TimeSeriesKeep 保留打点数据时间, 默认 DefaultTimeSeriesKeep
	TimeSeriesKeep *TimeSpan
	// TimeSeriesMaxSize 保留打点文件最大大小, 默认0表示不限制
	TimeSeriesMaxSize int64

	// RecycleCron 回收时间间隔 Cron 表达式，优先级比 RecycleInterval 高
	// 示例:
	// 午夜: @midnight 或 @daily
	// 每5分钟: @every 5m
	// 每时: @hourly
	// 每周: @weekly
	// 带秒的cron表达式:
	// 每秒: * * * * * ?
	// 每5分钟: 0 5 * * * *", every5min(time.Local)}
	// 每天2点: 0 0 2 * * ?
	// cron 表达式帮助: https://tool.lu/crontab/
	// cron 表达式帮助: https://www.bejson.com/othertools/cron/
	// 代码帮助: https://github.com/robfig/cron/blob/master/parser.go
	RecycleCron string
	// RecycleInterval 回收时间间隔, 默认 DefaultRecycleInterval
	RecycleInterval time.Duration
	// DividedBy 按时间分库模式
	DividedBy
}

const (
	// DefaultMaxIdle 默认最大数据库读写空闲时间，超期关闭数据库
	DefaultMaxIdle = 5 * time.Minute
	// DefaultBatchInterval 批量执行时间间隔
	DefaultBatchInterval = 10 * time.Second
	// DefaultBatchSize 批次大小
	DefaultBatchSize = 50

	// DefaultRecycleInterval 回收周期间隔
	DefaultRecycleInterval = 24 * time.Hour
)

const (
	DefaultKeySeqName = "keys.boltdb"
)

var (
	// DefaultTimeSeriesKeep 默认保留打点数据
	DefaultTimeSeriesKeep = TimeSpan{Value: 1, Unit: UnitMonth} // 1个月
)

// DefaultConfig 创建默认配置
func DefaultConfig() *Config {
	return &Config{
		DriverName:      "sqlite3",
		WriteDsnOptions: "_journal=WAL",
		ReadDsnOptions:  "_txlock=immediate",
		BadTableSubs:    []string{"-shm", "-wal", "-journal"},
		// ON CONFLICT clause does not match any PRIMARY KEY or UNIQUE constraint
		AllowDBErrors: []string{"no such", "has no column named", "does not match"},
		AsTags:        &noopFilter{},
	}
}

func createConfig(fns []ConfigFn) (*Config, error) {
	c := DefaultConfig()

	for _, f := range fns {
		f(c)
	}
	if c.Prefix != "" && !strings.HasSuffix(c.Prefix, ".") {
		c.Prefix += "."
	}
	if c.MaxIdle <= 0 {
		c.MaxIdle = DefaultMaxIdle
	}
	if c.BatchInsertInterval <= 0 {
		c.BatchInsertInterval = DefaultBatchInterval
	}
	if c.BatchInsertSize <= 0 {
		c.BatchInsertSize = DefaultBatchSize
	}
	if c.TimeSeriesKeep == nil {
		c.TimeSeriesKeep = &DefaultTimeSeriesKeep
	}
	if c.RecycleInterval <= 0 {
		c.RecycleInterval = DefaultRecycleInterval
	}

	var err error
	c.SeqKeysDB, c.SeqKeysDBName, err = CreateSeqKeysDB(c.Prefix, c.SeqKeysDBName, c.SeqKeysDB)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// CreateSeqKeysDB 创建	keys 字符串转换为枚举数字的 boltdb 库名字（可以包括路径），默认 keyseq.bolt
// 为 sqlite 的 tag 字符串值生成唯一的对应序号（减少sqlite数据库文件大小而优化设计）
//
//	prefix 设定库文件的前缀(包括完整路径）
//	seqKeysDBName 库名,  "off" 表示不使用, "" 使用 DefaultKeySeqName
//	seqKeysDB 外部已经提前创建好的库，在 seqKeysDBName != "off" 时优先使用
func CreateSeqKeysDB(prefix, seqKeysDBName string, seqKeysDB *BoltSeq) (*BoltSeq, string, error) {
	if seqKeysDBName == "off" {
		if seqKeysDB != nil {
			seqKeysDB.Close()
		}
		return nil, "off", nil
	}

	if seqKeysDB != nil {
		return seqKeysDB, "", nil
	}

	dir := filepath.Dir(prefix)
	if seqKeysDBName == "" {
		seqKeysDBName = filepath.Join(dir, DefaultKeySeqName)
	} else if !filepath.IsAbs(seqKeysDBName) {
		seqKeysDBName = filepath.Clean(filepath.Join(dir, seqKeysDBName))
	}

	boltdb, err := NewBoltSeq(seqKeysDBName, "keyseq")
	return boltdb, seqKeysDBName, err
}

// ValidateTable 校验表明是否合法
func (c *Config) ValidateTable(table string) error {
	for _, sub := range c.BadTableSubs {
		if strings.Contains(table, sub) {
			return fmt.Errorf("invalid table name: %s", table)
		}
	}

	return nil
}

// Matcher 匹配接口
type Matcher interface {
	Match(string) bool
}

// Sqliter Sqliter 结构体对象
type Sqliter struct {
	readDbsLock sync.Mutex
	readDbs     map[string]*readTable

	writeDbsLock sync.Mutex
	writeDbs     map[string]*writeTable

	*Config

	// recycleCancel 用于取消回收循环协程
	recycleCancel context.CancelFunc
}

type ConfigFn func(*Config)

func WithDebug(val bool) ConfigFn             { return func(c *Config) { c.Debug = val } }
func WithConfig(val *Config) ConfigFn         { return func(c *Config) { *c = *val } }
func WithDriverName(val string) ConfigFn      { return func(c *Config) { c.DriverName = val } }
func WithPrefix(val string) ConfigFn          { return func(c *Config) { c.Prefix = val } }
func WithDsnOptions(val string) ConfigFn      { return func(c *Config) { c.WriteDsnOptions = val } }
func WithBadTableSubs(val []string) ConfigFn  { return func(c *Config) { c.BadTableSubs = val } }
func WithAllowDBErrors(val []string) ConfigFn { return func(c *Config) { c.AllowDBErrors = val } }
func WithSeqKeysDBName(val string) ConfigFn   { return func(c *Config) { c.SeqKeysDBName = val } }
func WithSeqKeysDB(val *BoltSeq) ConfigFn     { return func(c *Config) { c.SeqKeysDB = val } }
func WithDividedBy(val DividedBy) ConfigFn    { return func(c *Config) { c.DividedBy = val } }
