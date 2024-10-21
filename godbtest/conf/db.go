package conf

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/godbtest/sqlmap"
	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/sqlparser"
	"github.com/bingoohuang/ngg/ss"
	"github.com/xo/dburl"
)

type DB interface {
	sqlmap.Queryer
	DBExecAware
}

type RunSQLOption struct {
	RowsScanner   sqlmap.RowsScanner
	Lookup        map[string]map[string]string
	Format        string
	RawFileDir    string
	RawFileExt    string
	Offset, Limit int
	DisplayMode   int
	Timeout       time.Duration // SQL 执行超时时间
	MaxLen        int
	PrintCost     bool
	SaveResult    bool
	TempResultDB  bool
	ShowRowIndex  bool
	NoEvalSQL     bool // 评估 SQL，提取常量为占位符，估算 @name 值等
	DryRun        bool
	ParsePrepared bool
	AsQuery       bool // 强制作为 query 执行（有结果集处理）
	AsExec        bool // 强制作为 更新执行（不处理结果集)
}

func (r *RunSQLOption) GetRowsScanner() sqlmap.RowsScanner {
	if r.RowsScanner != nil {
		return r.RowsScanner
	}

	return r.newRowsScanner()
}

func (r *RunSQLOption) newRowsScanner() sqlmap.RowsScanner {
	format := strings.ToLower(r.Format)
	vertical := r.DisplayMode == DisplayVertical
	switch {
	case r.DisplayMode == DisplayDefault && format == "none":
		return sqlmap.NewNoopRowsScanner()
	case r.DisplayMode == DisplayJSON || format == "json":
		return NewJsonRowsScanner(r.Offset, r.Limit, vertical, false)
	case r.DisplayMode == DisplayJSONFree || format == "json:free":
		return NewJsonRowsScanner(r.Offset, r.Limit, vertical, true)
	case format == "table":
		return NewTableRowsScanner("", r.Offset, r.Limit, vertical)
	case format == "markdown":
		return NewTableRowsScanner("markdown", r.Offset, r.Limit, vertical)
	case format == "csv":
		return NewTableRowsScanner("csv", r.Offset, r.Limit, vertical)
	case r.DisplayMode == DisplayInsertSQL || format == "insert":
		return NewInsertRowsScanner(r.Limit, r.Offset)
	default:
		return NewTableRowsScanner("", r.Offset, r.Limit, vertical)
	}
}

func RunSQL(ctx context.Context, db DB, driverName, q string, option RunSQLOption) error {
	dialectFn := genDialect(driverName)

	var (
		err  error
		args []any
	)

	if !option.NoEvalSQL {
		q, args, err = evalSQL(dialectFn, q, option.MaxLen, option.ParsePrepared)
		if err != nil {
			return err
		}
	}

	if option.DryRun {
		if option.NoEvalSQL {
			log.Printf("SQL: %q == %v", q, args)
		}
		return nil
	}

	scanner := option.GetRowsScanner()
	firstWord := strings.ToLower(strings.Fields(q)[0])

	options := []sqlmap.ScanConfigFn{
		sqlmap.WithShowRowIndex(option.ShowRowIndex),
		sqlmap.WithSaveResult(option.SaveResult),
		sqlmap.WithTempResultDB(option.TempResultDB),
		sqlmap.WithPrintCost(option.PrintCost),
		sqlmap.WithRawFile(option.RawFileDir, option.RawFileExt),
		sqlmap.WithLookup(option.Lookup),
	}

	if option.AsQuery {
		return Query(ctx, db, q, args, scanner, options...)
	} else if option.AsExec {
		return Exec(ctx, db, q, args, scanner)
	}

	switch firstWord {
	default:
		return Exec(ctx, db, q, args, scanner)
	case "select", "show", "desc", "describe", "call", "explain":
		return Query(ctx, db, q, args, scanner, options...)
	case "insert":
		if strings.Contains(strings.ToLower(q), "returning") {
			return Query(ctx, db, q, args, scanner, options...)
		}

		return Exec(ctx, db, q, args, scanner)
	}
}

func WithSignal(ctx context.Context, timeout time.Duration, sig ...os.Signal) (context.Context, context.CancelFunc) {
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}

	// trap Ctrl+C and call cancel on the context
	c := make(chan os.Signal, 1)
	if len(sig) == 0 {
		signal.Notify(c, os.Interrupt)
	} else {
		signal.Notify(c, sig...)
	}

	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx, func() {
		signal.Stop(c)
		cancel()
	}
}

func evalSQL(dialectFn BindNameAware, q0 string, maxLen int, parsePrepared bool) (q2 string, args []any, err error) {
	q := q0

	if parsePrepared {
		parsedQuery, subEvals, err := ParseSQL(dialectFn, q)
		if err != nil {
			return "", nil, err
		}

		if len(subEvals) == 0 {
			subEvals = fixBindPlaceholdersPrefix(parsedQuery, dialectFn)
		}

		if len(subEvals) > 0 {
			q = sqlparser.StringWithDialect(dialectFn.GetDialect(), parsedQuery) // insert ...
			args, err = genArgs(1, subEvals)
			if err != nil {
				return "", nil, err
			}
		}
	}

	ret, err := ss.ParseExpr(q).Eval(jj.NewSubstituter(substituteFns))
	if err != nil {
		return "", nil, err
	}

	q1 := ret.(string)
	if q != q1 {
		q = q1
	}

	if q != q0 || len(args) > 0 {
		if len(args) > 0 {
			abbreviateArgs := AbbreviateSlice(args, maxLen)
			log.Printf("SQL: %s ::: Args: %s", sqlmap.Color(q), ss.Json(abbreviateArgs))
		} else {
			log.Printf("SQL: %s;", sqlmap.Color(q))
		}
	}

	return q, args, nil
}

// ParseBool returns the boolean value represented by the string.
// It accepts 1, t, true, y, yes, on as true with camel case incentive
// and accepts 0, f false, n, no, off as false with camel case incentive
// Any other value returns an error.
func ParseBool(s string, defaultValue bool) bool {
	if s == "" {
		return defaultValue
	}
	switch strings.ToLower(s) {
	case "1", "t", "true", "y", "yes", "on":
		return true
	case "0", "f", "false", "n", "no", "off":
		return false
	}

	log.Panicf("unknown bool env value %q", s)
	return false
}

var UsingPing = ParseBool(os.Getenv("PING"), true)

func Query(ctx context.Context, db sqlmap.Queryer, q string, args []any, formatter sqlmap.RowsScanner, options ...sqlmap.ScanConfigFn) error {
	if UsingPing {
		if err := PingDB(ctx, db, 3*time.Second); err != nil {
			return err
		}
	}

	options1 := []sqlmap.ScanConfigFn{sqlmap.WithRowsScanner(formatter)}
	options1 = append(options1, options...)

	c := sqlmap.NewScanConfig(options1...)

	formatter.StartExecute(q)
	if err := c.Select(ctx, db, q, args...); err != nil {
		return fmt.Errorf("query %q failed: %w", q, err)
	}

	return nil
}

type DBExecAware interface {
	// ExecContext executes a query without returning any rows.
	// The args are for any placeholder parameters in the query.
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func Exec(ctx context.Context, db DBExecAware, q string, args []any, rowsScanner sqlmap.RowsScanner) error {
	if UsingPing {
		if err := PingDB(ctx, db, 3*time.Second); err != nil {
			return err
		}
	}

	start := time.Now()
	result, err := db.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("execute %q failed: %w", q, err)
	}
	cost := time.Since(start)

	id, err1 := result.LastInsertId()
	idStr := fmt.Sprintf("%d", id)
	if err1 != nil {
		idStr = "(N/A)"
	}
	affected, err2 := result.RowsAffected()
	affectedStr := fmt.Sprintf("%d", affected)
	if err2 != nil {
		affectedStr = "(N/A)"
	}
	rowsScanner.AddRow(0, []any{idStr, affectedStr})

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "lastInsertId: %s, rowsAffected: %s", idStr, affectedStr)
	fmt.Fprintf(&buf, ", cost: %s\n", cost)

	log.Printf("Result: %s", buf.String())
	return nil
}

func PingDB(ctx context.Context, db any, timeout time.Duration) error {
	ping, ok := db.(driver.Pinger)
	if !ok {
		return nil
	}

	timeoutCtx, cancelFunc := context.WithTimeout(ctx, timeout)
	defer cancelFunc()

	if err := ping.Ping(timeoutCtx); err == nil {
		return nil
	}

	// try again
	if err := ping.Ping(timeoutCtx); err != nil {
		log.Printf("PingDB failed: %v", err)
	}

	return nil
}

func init() {
	if scheme := dburl.Unregister("sqlite3"); scheme != nil {
		dburl.Unregister("sq")
		dburl.Unregister("file")

		scheme.Aliases = []string{"sq", "file"}
		dburl.Register(*scheme)
	}
	if scheme := dburl.Unregister("moderncsqlite"); scheme != nil {
		dburl.Unregister("sqlite")
		dburl.Unregister("modernsqlite")
		dburl.Unregister("mq")

		scheme.Driver = "sqlite"
		scheme.Aliases = []string{"mq", "modernsqlite"}
		dburl.Register(*scheme)
	}
}

func AutoTestConnect(dsnSource string) (driverName, fixedDataSourceName string) {
	u, err := dburl.Parse(dsnSource)
	if err != nil {
		log.Fatalf("parse datasource %v", err)
	}

	return u.Driver, u.DSN
}
