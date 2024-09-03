package rowscan

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bingoohuang/ngg/sqlparser"
	"github.com/bingoohuang/ngg/ss"
	"github.com/h2non/filetype"
	"github.com/samber/lo"
)

type ValueType int

const (
	_ ValueType = iota
	ValueTypeBool
	ValueTypeInt64
	ValueTypeFloat64
	ValueTypeString
	ValueTypeBytes
	ValueTypeOther
)

type RowScanner struct {
	Lookup map[string]map[string]string
	Rows   *sql.Rows

	RawFileDir   string
	RawFileExt   string
	Types        []ValueType
	LowerColumns []string
	Columns      []string
	SingleQuoted bool
}

type RowScannerFn func(*RowScanner)

func WithLookup(lookup map[string]map[string]string) RowScannerFn {
	return func(s *RowScanner) {
		s.Lookup = lookup
	}
}

func WithRawFile(dir, ext string) RowScannerFn {
	return func(o *RowScanner) {
		o.RawFileDir = dir
		o.RawFileExt = ext
	}
}

func WithSingleQuoted(singleQuoted bool) RowScannerFn {
	return func(o *RowScanner) {
		o.SingleQuoted = singleQuoted
	}
}

func NewRowScanner(query string, rows *sql.Rows, options ...RowScannerFn) (*RowScanner, error) {
	scanner := &RowScanner{
		Rows: rows,
	}
	for _, f := range options {
		f(scanner)
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	scanner.Columns = columns

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	scanner.Types = lo.Map(columnTypes, func(columnType *sql.ColumnType, index int) ValueType {
		switch typeName := strings.ToUpper(columnType.DatabaseTypeName()); {
		case ss.Contains(typeName, "CHAR", "TEXT", "NVARCHAR"):
			return ValueTypeString
		case ss.Contains(typeName, "BOOL"):
			return ValueTypeBool
		case ss.Contains(typeName, "BOOL", "INT"):
			return ValueTypeInt64
		case ss.Contains(typeName, "DECIMAL", "NUMBER"):
			return ValueTypeFloat64
		case ss.Contains(typeName, "LOB"):
			return ValueTypeBytes
		default:
			return ValueTypeOther
		}
	})

	if len(scanner.Lookup) > 0 {
		if lookupTable := strings.ToLower(GetSingleTableName(query)); lookupTable != "" {
			scanner.LowerColumns = lo.Map(columns, func(item string, idx int) string {
				return lookupTable + "." + strings.ToLower(item)
			})
		}
	}

	return scanner, nil
}

func (s RowScanner) Next() bool {
	return s.Rows.Next()
}

func (s RowScanner) Scan() ([]any, error) {
	values := lo.Map(s.Types, func(t ValueType, index int) any {
		return &NullAny{ValueType: t}
	})
	if err := s.Rows.Scan(values...); err != nil {
		return nil, err
	}

	row := lo.Map(values, func(val any, idx int) any {
		colValue := s.getColValue(*val.(*NullAny), s.SingleQuoted)
		colValue = s.lookup(idx, colValue)
		return colValue
	})
	return row, nil
}

func (s RowScanner) lookup(idx int, colValue any) any {
	if colValue == nil || len(s.LowerColumns) == 0 {
		return colValue
	}

	if m, ok := s.Lookup[s.LowerColumns[idx]]; ok {
		key := fmt.Sprintf("%v", colValue)
		if mapped, ok := m[key]; ok {
			colValue = mapped
		}
	}

	return colValue
}

func (s RowScanner) getColValue(n NullAny, singleQuoted bool) any {
	if !n.Valid {
		return nil
	}

	if n.ValueType == ValueTypeBytes {
		if data, ok := n.Value.([]byte); ok {
			ext := s.RawFileExt
			if ext == "" {
				if kind, _ := filetype.Match(data); kind.Extension != "" {
					ext = "." + kind.Extension
				}
			}

			dir := ss.Or(s.RawFileDir, ".tmp")
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				log.Printf("create %s: %v", dir, err)
				dir = ""
			}
			tempFileName, err := ss.WriteTempFile(dir, "*"+ext, data, false)
			if err != nil {
				log.Printf("write %s: %v", dir, err)
			}
			return "=> " + tempFileName
		}
	}

	switch n.ValueType {
	case ValueTypeInt64, ValueTypeFloat64:
		return n.Value
	}

	if !singleQuoted {
		return n.Value
	}

	return ss.QuoteSingle(fmt.Sprintf("%v", n.Value))
}

func GetSingleTableName(query string) string {
	result, err := sqlparser.ParseStrictDDL(query)
	if err != nil {
		return ""
	}

	sel, _ := result.(*sqlparser.Select)
	if sel == nil || len(sel.From) != 1 {
		return ""
	}

	expr, ok := sel.From[0].(*sqlparser.AliasedTableExpr)
	if !ok {
		return ""
	}

	return sqlparser.GetTableName(expr.Expr).String()
}
