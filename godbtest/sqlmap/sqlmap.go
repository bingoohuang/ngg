package sqlmap

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/godbtest/sqlmap/rowscan"
	"github.com/bingoohuang/ngg/ss"
	"github.com/samber/lo"
)

type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func Select(ctx context.Context, db Queryer, query string, args ...any) ([]map[string]any, error) {
	mapRowsScanner := NewMapRowsScanner()
	if err := NewScanConfig(WithRowsScanner(mapRowsScanner)).Select(ctx, db, query, args...); err != nil {
		return nil, err
	}

	return mapRowsScanner.Rows, nil
}

func (c *ScanConfig) Select(ctx context.Context, db Queryer, query string, args ...any) error {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	return c.Scan(query, rows)
}

type SingleQuoter interface {
	SingleQuote() bool
}

const (
	ShowRowIndex int = 1 << iota
	ShowCost
	SaveResult
)

type RowsScanner interface {
	StartExecute(query string)
	StartRows(query string, header []string, options int)
	// AddRow add a row cells value
	// If return true, the scanner will continue, or return false to stop scanning
	AddRow(rowIndex int, columns []any) bool
	Complete()
}

type noopRowsScanner struct {
	start time.Time
}

func (n *noopRowsScanner) StartExecute(string)             { n.start = time.Now() }
func (n *noopRowsScanner) StartRows(string, []string, int) {}
func (n noopRowsScanner) AddRow(int, []any) bool           { return true }
func (n *noopRowsScanner) Complete()                       { log.Printf("Cost %s", time.Since(n.start)) }

func NewNoopRowsScanner() RowsScanner { return &noopRowsScanner{} }

type MapRowsScanner struct {
	start  time.Time
	Header []string
	Rows   []map[string]any

	cost time.Duration
}

func NewMapRowsScanner() *MapRowsScanner {
	return &MapRowsScanner{
		Rows: []map[string]any{},
	}
}

func (n *MapRowsScanner) StartExecute(string) { n.start = time.Now() }
func (n *MapRowsScanner) StartRows(_ string, header []string, _ int) {
	n.Header = header
}

func (n *MapRowsScanner) AddRow(_ int, columns []any) bool {
	row := lo.SliceToMap(lo.Zip2(n.Header, columns),
		func(item lo.Tuple2[string, any]) (string, any) {
			return item.A, item.B
		})
	n.Rows = append(n.Rows, row)
	return true
}

func (n *MapRowsScanner) Complete() {
	n.cost = time.Since(n.start)
}

type ScanConfig struct {
	RowsScanner

	Lookup map[string]map[string]string

	RawFileDir   string
	RawFileExt   string
	ShowRowIndex bool
	PrintCost    bool
	SaveResult   bool

	// 是否将执行结果放到 sqlite 表中
	TempResultDB bool
}

type ScanConfigFn func(*ScanConfig)

func WithShowRowIndex(value bool) ScanConfigFn {
	return func(o *ScanConfig) { o.ShowRowIndex = value }
}

func WithTempResultDB(value bool) ScanConfigFn {
	return func(o *ScanConfig) { o.TempResultDB = value }
}

func WithSaveResult(value bool) ScanConfigFn {
	return func(o *ScanConfig) { o.SaveResult = value }
}

func WithPrintCost(value bool) ScanConfigFn {
	return func(o *ScanConfig) { o.PrintCost = value }
}

func WithLookup(lookup map[string]map[string]string) ScanConfigFn {
	return func(o *ScanConfig) {
		o.Lookup = lookup
	}
}

func WithRawFile(dir, ext string) ScanConfigFn {
	return func(o *ScanConfig) {
		o.RawFileDir = dir
		o.RawFileExt = ext
	}
}

func WithRowsScanner(value RowsScanner) ScanConfigFn {
	return func(o *ScanConfig) { o.RowsScanner = value }
}

func Scan(query string, rows *sql.Rows) ([]map[string]any, error) {
	mapRowsScanner := NewMapRowsScanner()
	if err := ScanOption(query, rows, WithRowsScanner(mapRowsScanner)); err != nil {
		return nil, err
	}

	return mapRowsScanner.Rows, nil
}

func ScanOption(query string, rows *sql.Rows, options ...ScanConfigFn) error {
	return NewScanConfig(options...).Scan(query, rows)
}

func crateResultDB() (resultDBName string, resultDB *sql.DB) {
	resultDBName = ".r" + time.Now().Format(`20060102`) + ".db"
	resultDB, _ = sql.Open("sqlite", resultDBName)
	return resultDBName, resultDB
}

func (c *ScanConfig) Scan(query string, rows *sql.Rows) error {
	defer ss.Close(rows)

	singleQuoted := false
	if q, ok := c.RowsScanner.(SingleQuoter); ok {
		singleQuoted = q.SingleQuote()
	}

	scan, err := rowscan.NewRowScanner(query, rows, rowscan.WithLookup(c.Lookup),
		rowscan.WithRawFile(c.RawFileDir, c.RawFileExt),
		rowscan.WithSingleQuoted(singleQuoted))
	if err != nil {
		return err
	}

	options := 0
	options = BitOr(c.ShowRowIndex, options, ShowRowIndex)
	options = BitOr(c.PrintCost, options, ShowCost)
	options = BitOr(c.SaveResult, options, SaveResult)

	tempResultSQL := ""
	tempTable := `r` + time.Now().Format(`150405`)
	var insertTempDbError error
	var resultDBName string
	var resultDB *sql.DB

	if c.TempResultDB {
		resultDBName, resultDB = crateResultDB()
		if resultDB != nil {
			defer resultDB.Close()
		}

		ddl := `create table ` + tempTable
		for i, col := range scan.Columns {
			ddl += ss.If(i == 0, `(`, `,`) + col + ` text`
		}
		ddl += `)`
		if _, err := resultDB.Exec(ddl); err != nil {
			log.Printf("exec %s error: %v", ddl, err)
		}

		tempResultSQL = fmt.Sprintf("insert into %s(%s) values(%s)", tempTable,
			strings.Join(scan.Columns, ", "), strings.Repeat(",?", len(scan.Columns))[1:])

		log.Printf("result will be saved to db: %s table: %s", resultDBName, tempTable)
	}

	c.RowsScanner.StartRows(query, scan.Columns, options)
	rowNum := 0
	for ; scan.Next(); rowNum++ {
		row, err := scan.Scan()
		if err != nil {
			return err
		}

		if c.TempResultDB {
			if _, err := resultDB.Exec(tempResultSQL, row...); err != nil {
				insertTempDbError = err
			}
		}

		if !c.RowsScanner.AddRow(rowNum, row) {
			break
		}
	}

	c.RowsScanner.Complete()

	if c.TempResultDB {
		if insertTempDbError != nil {
			log.Printf("exec %s error: %v", tempResultSQL, err)
		} else {
			log.Printf("#%d rows saved to db: %s table: %s", rowNum, resultDBName, tempTable)
		}
	}

	return rows.Err()
}

func NewScanConfig(options ...ScanConfigFn) *ScanConfig {
	conf := &ScanConfig{}
	for _, f := range options {
		f(conf)
	}

	if conf.RowsScanner == nil {
		conf.RowsScanner = NewNoopRowsScanner()
	}
	return conf
}

func BitOr(condition bool, a, b int) int {
	if condition {
		return a | b
	}

	return a
}
