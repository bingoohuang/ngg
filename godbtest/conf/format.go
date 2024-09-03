package conf

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/godbtest/sqlmap"
	"github.com/bingoohuang/ngg/godbtest/sqlmap/rowscan"
	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/ss"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/samber/lo"
)

func NewInsertRowsScanner(limit, offset int) *InsertRowsScanner {
	return &InsertRowsScanner{
		Limit:  limit,
		Offset: offset,
	}
}

type InsertRowsScanner struct {
	sqlFile   *os.File
	tableName string

	Header        []string
	Limit, Offset int
	rows          int
}

func (j *InsertRowsScanner) SingleQuote() bool { return true }

func (j *InsertRowsScanner) StartExecute(query string) {
}

func (j *InsertRowsScanner) StartRows(query string, header []string, options int) {
	j.Header = header
	j.tableName = ss.Or(rowscan.GetSingleTableName(query), "table_name")
	sqlFile, err := os.CreateTemp("", "*.sql")
	if err != nil {
		log.Printf("create temp file failed: %v", err)
	}

	if options&sqlmap.SaveResult == sqlmap.SaveResult {
		j.sqlFile = sqlFile
	}
}

func (j *InsertRowsScanner) AddRow(rowIndex int, columns []any) bool {
	if j.Offset > 0 && rowIndex < j.Offset {
		return true
	}

	var fields, quoted []string

	for i, c := range columns {
		if f, ok := c.(float64); ok {
			c = FormatFloat64(f)
		}
		if c != nil {
			fields = append(fields, j.Header[i])
			quoted = append(quoted, fmt.Sprintf("%v", c))
		}
	}

	insertSQL := fmt.Sprintf("insert into %s(%s) values(%s);", j.tableName,
		strings.Join(fields, ", "), strings.Join(quoted, ", "))

	if j.sqlFile != nil {
		_, _ = j.sqlFile.WriteString(insertSQL + "\n")
	}

	fmt.Println(insertSQL)
	j.rows++
	return j.Limit <= 0 || rowIndex+1 <= j.Limit+j.Offset
}

func (j *InsertRowsScanner) Complete() {
	if j.sqlFile != nil {
		_ = j.sqlFile.Close()
		if j.rows > 0 {
			log.Printf("result saved to %s", j.sqlFile.Name())
		} else {
			_ = os.Remove(j.sqlFile.Name())
		}
	}
}

type JsonRowsScanner struct {
	start         time.Time
	jsonFile      *os.File
	Header        []string
	Limit, Offset int
	options       int
	rows          int
	Vertical      bool
	FreeInnerJSON bool
	showRowIndex  bool
}

func NewJsonRowsScanner(offset, limit int, vertical, freeInnerJSON bool) *JsonRowsScanner {
	return &JsonRowsScanner{Limit: limit, Offset: offset, Vertical: vertical, FreeInnerJSON: freeInnerJSON}
}

var _ sqlmap.RowsScanner = (*JsonRowsScanner)(nil)

func (j *JsonRowsScanner) StartExecute(string) {
	j.start = time.Now()
}

func (j *JsonRowsScanner) StartRows(_ string, header []string, options int) {
	j.Header = header
	j.options = options
	j.showRowIndex = options&sqlmap.ShowRowIndex == sqlmap.ShowRowIndex
	if j.options&sqlmap.SaveResult == sqlmap.SaveResult {
		j.jsonFile, _ = os.CreateTemp("", "*.json")
	}
}

func (j *JsonRowsScanner) AddRow(rowIndex int, columns []any) bool {
	if j.Offset > 0 && rowIndex < j.Offset {
		return true
	}

	if rowIndex+1 > j.Limit+j.Offset {
		log.Printf("Row ignored ...")
		return false
	}

	row := map[string]any{}
	for i, h := range j.Header {
		if columns[i] != nil {
			row[h] = columns[i]
		}
	}

	rowJSON, _ := json.Marshal(row)
	if j.FreeInnerJSON {
		rowJSON = jj.FreeInnerJSON(rowJSON)
	}

	if j.Vertical {
		rowJSON = jj.Pretty(rowJSON)
	}

	if j.jsonFile != nil {
		_, _ = j.jsonFile.Write(rowJSON)
		_, _ = j.jsonFile.WriteString("\n")
	}
	j.rows++

	colorJSON := jj.Color(rowJSON, nil, nil)

	if j.showRowIndex {
		fmt.Printf("Row %03d: %s\n", rowIndex+1, colorJSON)
	} else {
		fmt.Printf("%s\n", colorJSON)
	}

	return j.Limit <= 0 || rowIndex+1 <= j.Limit+j.Offset
}

func (j *JsonRowsScanner) Complete() {
	if j.options&sqlmap.ShowCost == sqlmap.ShowCost {
		log.Printf("Cost %s", time.Since(j.start))
	}

	if j.jsonFile != nil {
		_ = j.jsonFile.Close()
		if j.rows > 0 {
			log.Printf("result saved to %s", j.jsonFile.Name())
		} else {
			_ = os.Remove(j.jsonFile.Name())
		}
	}
}

type TableRowsScanner struct {
	start  time.Time
	Table  table.Writer
	Format string
	Header []string

	Limit, Offset int
	options       int
	rows          int
	RowVertical   bool
	showRowIndex  bool
}

func NewTableRowsScanner(format string, offset, limit int, rowVertical bool) *TableRowsScanner {
	t := &TableRowsScanner{
		Format:      format,
		Limit:       limit,
		Offset:      offset,
		RowVertical: rowVertical,
	}
	if !rowVertical {
		t.Table = createTableWriter()
	}
	return t
}

func createTableWriter() table.Writer {
	t := table.NewWriter()
	style := table.StyleDefault
	style.Format.Header = text.FormatDefault
	t.SetStyle(style)

	return t
}

func (t *TableRowsScanner) StartExecute(string) { t.start = time.Now() }

func (t *TableRowsScanner) StartRows(_ string, header []string, options int) {
	t.options = options
	if t.RowVertical {
		t.Header = header
		return
	}

	headers := lo.Map(header, func(item string, index int) any { return item })
	t.showRowIndex = t.options&sqlmap.ShowRowIndex == sqlmap.ShowRowIndex

	if t.showRowIndex {
		headers = append(table.Row{"#"}, headers...)
	}
	t.Table.AppendHeader(headers)
}

func (t *TableRowsScanner) AddRow(rowIndex int, columns []any) bool {
	for i, col := range columns {
		if f, ok := col.(float64); ok {
			columns[i] = FormatFloat64(f)
		}
	}

	if t.RowVertical {
		if t.Offset > 0 && rowIndex < t.Offset {
			return true
		}

		tw := createTableWriter()
		if t.showRowIndex {
			tw.AppendHeader([]any{"#", "Column", fmt.Sprintf("value of row %d", rowIndex+1)})
		} else {
			tw.AppendHeader([]any{"Column", fmt.Sprintf("value of row %d", rowIndex+1)})
		}
		for i, h := range t.Header {
			if t.showRowIndex {
				tw.AppendRow([]any{i + 1, h, columns[i]})
			} else {
				tw.AppendRow([]any{h, columns[i]})
			}
		}
		content, extension := t.render(tw)
		fmt.Println(content)

		if t.options&sqlmap.SaveResult == sqlmap.SaveResult {
			writeTempFile(content, extension)
		}

		return rowIndex+1 <= t.Limit+t.Offset
	}

	if t.Limit > 0 && rowIndex+1 > t.Limit+t.Offset {
		t.Table.AppendRow([]any{"..."})
		return false
	}

	t.rows++

	if t.Offset > 0 && rowIndex < t.Offset {
		return true
	}

	if t.showRowIndex {
		columns = append([]any{rowIndex + 1}, columns...)
	}

	t.Table.AppendRow(columns)

	return t.Limit <= 0 || rowIndex+1 <= t.Limit+t.Offset
}

// FormatFloat64 formats a float64 number to a string with the specified precision.
// If the fractional part is all zeros, it returns the integer part only.
func FormatFloat64(f float64) string {
	// The -1 as the third parameter tells the function to print
	// the fewest digits necessary to accurately represent the float.
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// isTrailingZeros checks if the given string consists only of zeros.
func isTrailingZeros(s string) bool {
	for _, ch := range s {
		if ch != '0' {
			return false
		}
	}
	return true
}

func writeTempFile(content, extension string) {
	if temp, _ := os.CreateTemp("", "*"+extension); temp != nil {
		_, _ = temp.WriteString(content + "\n")
		_ = temp.Close()
		log.Printf("result saved to %s", temp.Name())
	}
}

func (t TableRowsScanner) Complete() {
	if t.options&sqlmap.ShowCost == sqlmap.ShowCost {
		defer log.Printf("Cost %s", time.Since(t.start))
	}
	if t.RowVertical {
		return
	}

	content, extension := t.render(t.Table)
	fmt.Println(content)

	if t.rows > 0 && t.options&sqlmap.SaveResult == sqlmap.SaveResult {
		writeTempFile(content, extension)
	}
}

func (t TableRowsScanner) render(table table.Writer) (content, extension string) {
	switch t.Format {
	case "markdown":
		return table.RenderMarkdown(), ".md"
	case "csv":
		return table.RenderCSV(), ".csv"
	case "html":
		return table.RenderHTML(), ".html"
	default:
		return table.Render(), ".txt"
	}
}

var _ sqlmap.RowsScanner = (*TableRowsScanner)(nil)
