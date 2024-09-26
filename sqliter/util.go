package sqliter

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bingoohuang/ngg/tick"

	"github.com/bingoohuang/ngg/sqlrun"
	"github.com/bingoohuang/ngg/ss"
	pie "github.com/elliotchance/pie/v2"
	"github.com/samber/lo"
)

func SortMap(m map[string]any) (keys []string, values []any) {
	keys = lo.Keys(m)
	sort.Strings(keys)

	values = pie.Map(keys, func(k string) any { return m[k] })
	return keys, values
}

// OnConflictGen 生成 on conflict 子语句函数
func OnConflictGen(tags map[string]bool) func(columns []string) string {
	q := " on conflict(" + strings.Join(pie.Map(lo.Keys(tags), strconv.Quote), ",") + ") do update set "
	return func(columns []string) string {
		sets := pie.Of(columns).
			Filter(func(s string) bool {
				return !tags[s]
			}).
			Map(func(s string) string {
				return s + "=excluded." + s
			}).
			Result
		return q + strings.Join(sets, ",")
	}
}

func ToDBFieldValue(x any) any {
	if x == nil {
		return nil
	}

	switch v := x.(type) {
	case string:
		return v
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return uint64(v)
	case uint8:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint32:
		return uint64(v)
	case uint64:
		return v
	case float32:
		return float64(v)
	case float64:
		return v
	case time.Duration:
		return v.String()
	case time.Time:
		// Universal Coordinated Time (UTC) is used.
		// https://www.sqlite.org/lang_datefunc.html
		return v.Format(`2006-01-02 15:04:05.000`)
	default:
		if s, ok := x.(fmt.Stringer); ok {
			return s.String()
		}

		vv, _ := json.Marshal(x)
		return vv
	}
}

func Tick[T any](interval, jitter time.Duration, ch chan T, chItemFn func(T) error, tickFn func() error) error {
	t := time.NewTimer(tick.Jitter(interval, jitter))
	defer t.Stop()

	if tickFn == nil {
		tickFn = func() error { return nil }
	}
	for {
		select {
		case item, ok := <-ch:
			if !ok {
				return tickFn()
			}
			if err := chItemFn(item); err != nil {
				return err
			}
		case <-t.C:
			if err := tickFn(); err != nil {
				return err
			}
			t.Reset(tick.Jitter(interval, jitter))
		}
	}
}

func MultiInsertBinds(columnsNum, rowsNum int) string {
	rowBinds := "(" + ss.Repeat("?", ",", columnsNum) + ")"
	return ss.Repeat(rowBinds, ", ", rowsNum)
}

func QuoteSlice(ss []string) []string {
	quoted := make([]string, len(ss))

	for i, column := range ss {
		quoted[i] = strconv.Quote(column)
	}

	return quoted
}

type DebugDB struct {
	// lock 只有在执行更新时使用，查询不使用
	lock sync.Mutex
	DB   *sql.DB

	// DSN 连接字符串
	DSN string

	*Config
	maxOpenConns int
}

func NewDebugDB(dsn string, maxOpenConns int, c *Config) (*DebugDB, error) {
	db, err := sql.Open(c.DriverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("openDB, driver: %s, dsn: %s: %v", c.DriverName, dsn, err)
	}
	db.SetMaxOpenConns(maxOpenConns)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("pingDB, driver: %s, dsn: %s: %v", c.DriverName, dsn, err)
	}

	return &DebugDB{
		maxOpenConns: maxOpenConns,
		DB:           db,
		DSN:          dsn,
		Config:       c,
	}, nil
}

func (d *DebugDB) Close() {
	if err := d.DB.Close(); err != nil && d.Debug {
		log.Printf("E! closeDB: %s, error: %v", d.DSN, err)
	}
}

// Query 执行查询
// emptyStruct 是 nil 时, result 中 Rows 格式是 [][]string
// emptyStruct 非 nil 时, result 中 Rows 格式是 []MyStruct, emptyStruct 传入例如 MyStruct{} 的空结构体实例
func (d *DebugDB) Query(emptyStruct any, query string, vars ...any) (*sqlrun.Result, error) {
	r, err := sqlrun.SetResultType(emptyStruct).Query(d.DB, query, vars...)
	if err != nil {
		log.Printf("E! dbQuery: %s, args: %v, DSN: %v, error: %v", query, vars, d.DSN, err)
		return nil, err
	}

	if d.Debug {
		log.Printf("dbQuery: %s, args: %v, DSN: %v, cost: %v, rows: %d", query, vars, d.DSN, r.CostTime, r.RowsCount)
	}
	return r, nil
}

func (d *DebugDB) dbExec(dbLocked bool, q string, args ...any) *sqlrun.Result {
	if !dbLocked {
		d.lock.Lock()
		defer d.lock.Unlock()
	}

	r, err := sqlrun.Query(d.DB, q, args...)
	if err != nil {
		log.Printf("E! dbExec: %s, args: %v, error: %v", q, args, err)
		return r
	} else if d.Debug {
		log.Printf("dbExec: %s, args: %v, cost: %s, rowsAffected: %d", q, args, r.CostTime, r.RowsAffected)
	}
	return r
}

// Prepared 预备语句
type Prepared struct {
	// batchSize 批量插入大小
	batchSize int

	query string
	db    *DebugDB
	stmt  *sql.Stmt
	Debug bool

	err error
}

func (d *Prepared) Close() error {
	return d.stmt.Close()
}

func (d *Prepared) Exec(args ...any) (sql.Result, error) {
	d.db.lock.Lock()
	defer d.db.lock.Unlock()

	result, err := d.stmt.Exec(args...)
	return result, err

}

const tooManySqlVariables = "too many SQL variables"

// 预备语句
// 可能产生错误 err:too many SQL variables
// SQLite数据库在执行SQL语句时，对变量的数量有限制，这个限制称为SQLITE_MAX_VARIABLE_NUMBER。
// 默认情况下，这个限制在SQLite版本 3.32.0 之前的版本是 999，而在 3.32.0 及之后的版本是 32766
// github.com/mattn/go-sqlite3 v2.0.3+incompatible 时， SELECT sqlite_version() 结果 3.31.1, 最大变量数量是 999
// github.com/mattn/go-sqlite3 v1.14.22 时,  SELECT sqlite_version() 结果 3.45.1, 最大变量数量是 32766
func (d *DebugDB) dbPrepare(baseQuery string, batchSize int, postfix func(batchSize int) string, logErr bool) *Prepared {
	d.lock.Lock()
	defer d.lock.Unlock()

	var (
		err  error
		stmt *sql.Stmt
		cost time.Duration
		q    string
	)

	newBatchSize := batchSize
	for {
		q = baseQuery + postfix(newBatchSize)
		start := time.Now()
		stmt, err = d.DB.Prepare(q)
		cost = time.Since(start)
		if err != nil {
			if strings.Contains(err.Error(), tooManySqlVariables) {
				newBatchSize--
				continue
			}
		}
		break
	}

	if err != nil {
		if logErr {
			log.Printf("E! dbPrepare: %s, error: %v", q, err)
		}
	} else if d.Debug {
		log.Printf("dbPrepare: %s, cost: %s", q, cost)
	}

	if newBatchSize != batchSize {
		log.Printf("W! newBatchSize: %d", newBatchSize)
	}

	return &Prepared{
		query:     q,
		db:        d,
		stmt:      stmt,
		Debug:     d.Debug,
		batchSize: newBatchSize,
		err:       err,
	}
}

type file struct {
	size int64
	path string
}

// RemoveFilesPrefix 删除以特定前缀开头的文件
func RemoveFilesPrefix(prefix string, debug bool) (removeFiles []file, totalSize int64) {
	if stat, err := os.Stat(prefix); err != nil || stat.IsDir() {
		return
	}

	dirPath := filepath.Dir(prefix)
	err := filepath.WalkDir(dirPath, func(path string, info os.DirEntry, err error) error {
		if !info.IsDir() && strings.HasPrefix(path, prefix) {
			f := file{path: path}
			if fi, err := info.Info(); err == nil {
				f.size = fi.Size()
			}
			removeFiles = append(removeFiles, f)
		}
		return err
	})

	if err != nil {
		log.Printf("E! walkdir: %s error: %v", dirPath, err)
	}

	for _, f := range removeFiles {
		if e := os.Remove(f.path); e != nil {
			log.Printf("E! removeFile: %s, error: %v", f.path, e)
		} else {
			totalSize += f.size
			if debug {
				log.Printf("removeFile: %s, size: %s", f.path, ss.IBytes(uint64(f.size)))
			}
		}
	}

	return
}

// https://www.sqlite.org/datatype3.html
// Each value stored in an SQLite database (or manipulated by the database engine)
// has one of the following storage classes:
// NULL. The value is a NULL value.
// INTEGER. The value is a signed integer, stored in 1, 2, 3, 4, 6, or 8 bytes depending on the magnitude of the value.
// REAL. The value is a floating point value, stored as an 8-byte IEEE floating point number.
// TEXT. The value is a text string, stored using the database encoding (UTF-8, UTF-16BE or UTF-16LE).
// BLOB. The value is a blob of data, stored exactly as it was input.

// https://www.sqlite.org/datatype3.html
// SQLite does not have a storage class set aside for storing dates and/or times.
// Instead, the built-in Date And Time Functions of SQLite are capable of storing dates and times
// as TEXT, REAL, or INTEGER values:
//
// TEXT as ISO8601 strings ("YYYY-MM-DD HH:MM:SS.SSS").
// REAL as Julian day numbers, the number of days since noon in Greenwich on November 24, 4714 B.C.
// according to the proleptic Gregorian calendar.
// INTEGER as Unix Time, the number of seconds since 1970-01-01 00:00:00 UTC.
func deriveDataType(columnName string, value any, asTags Matcher, seq *BoltSeq) string {
	switch value.(type) {
	// 当文本数据插入到 NUMERIC 列中时，如果文本分别是格式良好的整数或实值，则文本的存储类将转换为 INTEGER 或 REAL (按优先顺序)。
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return "NUMERIC"
	// 具有 TEXT 亲和类型的列使用 NULL、 TEXT 或 BLOB 存储类存储所有数据。
	default:
		if seq != nil && asTags.Match(columnName) {
			return "NUMERIC"
		}
		return "TEXT"
	}
}

var timeReg = regexp.MustCompile(`^(-?\d+)([hms])$`)

func ParseTimeWindow(timeWindow string) (window time.Duration, err error) {
	sub := timeReg.FindStringSubmatch(timeWindow)
	if len(sub) != 3 {
		return 0, errors.New("bad time")
	}

	val, _ := strconv.ParseInt(sub[1], 10, 64)
	switch sub[2] {
	case "d":
		return time.Duration(val) * 24 * time.Hour, nil
	case "h":
		return time.Duration(val) * time.Hour, nil
	case "m":
		return time.Duration(val) * time.Minute, nil
	case "s":
		return time.Duration(val) * time.Second, nil
	default:
		return time.Duration(val), nil
	}
}

type noopFilter struct{}

func (n noopFilter) Match(string) bool { return false }
