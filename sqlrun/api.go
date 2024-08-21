package sqlrun

import "database/sql"

type run struct {
	emptyStruct any
	maxRows     int
}

type Run interface {
	SetMaxRows(maxRows int) Run
	SetResultType(emptyStruct any) Run

	Query(db miniDB, query string, vars ...any) (*Result, error)
}

func SetMaxRows(maxRows int) Run {
	var r run
	return r.SetMaxRows(maxRows)
}
func SetResultType(emptyStruct any) Run {
	var r run
	return r.SetResultType(emptyStruct)
}

func Query(db miniDB, query string, vars ...any) (*Result, error) {
	var r run
	return r.Query(db, query, vars...)
}

func (r run) SetMaxRows(maxRows int) Run {
	r.maxRows = maxRows
	return r
}

func (r run) SetResultType(emptyStruct any) Run {
	r.emptyStruct = emptyStruct
	return r
}

// Query 执行查询
// 没有设置 SetResultType 时, result 中 Rows 格式是 [][]string
// 已经设置 SetResultType(MyStruct{}) 时, result 中 Rows 格式是 []MyStruct
func (r run) Query(db miniDB, query string, vars ...any) (*Result, error) {
	var fns []OptionFn
	if r.emptyStruct != nil {
		fns = append(fns, WithRowsPrepare(newStructPreparer(r.emptyStruct)))
	}
	if r.maxRows > 0 {
		fns = append(fns, WithMaxRows(r.maxRows))
	}

	result := newSQLRun(db, fns...).DoExec(query, vars...)
	if result.Error != nil {
		return nil, result.Error
	}

	return result, nil
}

type miniDB interface {
	// Exec executes update.
	Exec(query string, args ...any) (sql.Result, error)
	// Query performs query.
	Query(query string, args ...any) (*sql.Rows, error)
}
