package sqlrun

import (
	"database/sql"

	"github.com/bingoohuang/ngg/ss"
)

// newStrPreparer creates a new strPreparer.
func newStrPreparer(nullReplace string) *strPreparer {
	return &strPreparer{
		NullReplace: nullReplace,
	}
}

// strPreparer prepares to scan query rows.
type strPreparer struct {
	// NullReplace is the replacement of null values.
	NullReplace string
}

// Prepare prepares to scan query rows.
func (m *strPreparer) Prepare(rows *sql.Rows, columns []string) mapping {
	columnSize := len(columns)
	columnTypes, _ := rows.ColumnTypes()
	columnLobs := make([]bool, columnSize)

	for i := 0; i < columnSize; i++ {
		columnLobs[i] = ContainsFold(columnTypes[i].DatabaseTypeName(), "LOB")
	}

	return &MapMapping{
		columnSize:  columnSize,
		nullReplace: m.NullReplace,
		columnTypes: columnTypes,
		columnLobs:  columnLobs,
		rows:        rows,
		rowsData:    make([][]string, 0),
	}
}

// MapMapping maps the query rows to maps.
type MapMapping struct {
	columnSize  int
	nullReplace string
	columnLobs  []bool
	columnTypes []*sql.ColumnType
	rows        *sql.Rows
	rowsData    [][]string
}

// RowsData returns the mapped rows data.
func (m *MapMapping) RowsData() any { return m.rowsData }

// Scan scans the rows one by one.
func (m *MapMapping) Scan(rowNum int) error {
	holders := make([]sql.NullString, m.columnSize)
	pointers := make([]any, m.columnSize)

	for i := 0; i < m.columnSize; i++ {
		pointers[i] = &holders[i]
	}

	err := m.rows.Scan(pointers...)
	if err != nil {
		return err
	}

	values := make([]string, m.columnSize)

	for i, h := range holders {
		values[i] = ss.If(h.Valid, h.String, m.nullReplace)

		if h.Valid && m.columnLobs[i] {
			values[i] = "(" + m.columnTypes[i].DatabaseTypeName() + ")"
		}
	}

	m.rowsData = append(m.rowsData, values)
	return nil
}
