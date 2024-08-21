package sqlrun

import (
	"strings"
	"time"
)

// Result defines the result structure of sql execution.
type Result struct {
	Error    error
	CostTime time.Duration

	Headers   []string
	Rows      any // [][]string or []YourStruct
	RowsCount int

	RowsAffected int64
	LastInsertID int64
	IsQuery      bool
}

// StringRows return the string rows when using strPreparer.
func (r Result) StringRows() [][]string {
	return r.Rows.([][]string)
}

// SQLRun is used to execute queries and updates.
type SQLRun struct {
	miniDB

	RowsPrepare // required only for query

	MaxRows int
}

type OptionFn func(run *SQLRun)

func WithRowsPrepare(row RowsPrepare) OptionFn {
	return func(run *SQLRun) {
		run.RowsPrepare = row
	}
}
func WithMaxRows(maxRows int) OptionFn {
	return func(run *SQLRun) {
		run.MaxRows = maxRows
	}
}

// newSQLRun creates a new SQLRun for queries and updates.
func newSQLRun(db miniDB, fns ...OptionFn) *SQLRun {
	r := &SQLRun{miniDB: db}
	for _, fn := range fns {
		fn(r)
	}

	return r
}

// DoExec executes a SQL.
func (s *SQLRun) DoExec(query string, args ...any) *Result {
	_, isQuerySQL := IsQuerySQL(query)
	if isQuerySQL {
		return s.DoQuery(query, args...)
	}

	return s.DoUpdate(query, args...)
}

// DoUpdate does the update.
func (s *SQLRun) DoUpdate(query string, vars ...any) *Result {
	result := &Result{}
	start := time.Now()
	r, err := s.Exec(query, vars...)

	if r != nil {
		result.RowsAffected, _ = r.RowsAffected()
		result.LastInsertID, _ = r.LastInsertId()
	}

	result.Error = err
	result.CostTime = time.Since(start)

	return result
}

// DoQuery does the query.
func (s *SQLRun) DoQuery(query string, args ...any) *Result {
	result := &Result{}
	start := time.Now()
	result.IsQuery = true

	defer func() {
		result.CostTime = time.Since(start)
	}()

	rows, err := s.Query(query, args...)
	if err != nil {
		result.Error = err
		return result
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		result.Error = err
		return result
	}

	rowsPrepare := s.RowsPrepare
	if rowsPrepare == nil {
		rowsPrepare = newStrPreparer("")
	}
	mapping := rowsPrepare.Prepare(rows, columns)

	r := 0
	for ; rows.Next() && (s.MaxRows <= 0 || r < s.MaxRows); r++ {
		if err := mapping.Scan(r); err != nil {
			result.Error = err
			return result
		}
	}

	result.Error = err
	result.Headers = columns
	result.Rows = mapping.RowsData()
	result.RowsCount = r
	return result
}

// IsQuerySQL tests a sql is a query or not.
func IsQuerySQL(sql string) (string, bool) {
	key := FirstWord(sql)

	switch strings.ToUpper(key) {
	case "INSERT", "DELETE", "UPDATE", "SET", "REPLACE":
		return key, false
	case "SELECT", "SHOW", "DESC", "DESCRIBE", "EXPLAIN":
		return key, true
	default:
		return key, false
	}
}

// FirstWord returns the first word of the SQL statement s.
func FirstWord(s string) string {
	if fields := strings.Fields(strings.TrimSpace(s)); len(fields) > 0 {
		return fields[0]
	}

	return ""
}
