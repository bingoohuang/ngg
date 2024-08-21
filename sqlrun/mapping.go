package sqlrun

import (
	"database/sql"
	"reflect"
	"strings"

	"github.com/bingoohuang/ngg/ss"
)

type selectItem interface {
	Type() reflect.Type
	SetField(val reflect.Value)
	SetRoot(root reflect.Value)
}

type structItem struct {
	field *reflect.StructField
	root  reflect.Value
}

func (s *structItem) Type() reflect.Type         { return s.field.Type }
func (s *structItem) SetRoot(root reflect.Value) { s.root = root }
func (s *structItem) SetField(val reflect.Value) {
	s.root.FieldByIndex(s.field.Index).Set(val.Convert(s.field.Type))
}

// mapping defines the interface for SQL query processing.
type mapping interface {
	Scan(rowNum int) error
	RowsData() any
}

// RowsPrepare prepares to scan query rows.
type RowsPrepare interface {
	// Prepare prepares to scan query rows.
	Prepare(rows *sql.Rows, columns []string) mapping
}

// implType tells src whether it implements target type.
func implType(src, target reflect.Type) bool {
	if src == target {
		return true
	}

	if src.Kind() == reflect.Ptr {
		return src.Implements(target)
	}

	if target.Kind() != reflect.Interface {
		return false
	}

	return reflect.PointerTo(src).Implements(target)
}

// 参考 https://github.com/uber-go/dig/blob/master/types.go
var (
	_sqlScannerType = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
)

// implSQLScanner tells t whether it implements sql.Scanner interface.
func implSQLScanner(t reflect.Type) bool { return implType(t, _sqlScannerType) }

type selectItemSlice []selectItem

// newStructFields creates new struct fields slice.
func (m *structPreparer) newStructFields(columns []string) selectItemSlice {
	mapFields := make(selectItemSlice, len(columns))
	for i, col := range columns {
		mapFields[i] = m.newStructField(col)
	}

	return mapFields
}

// newStructField creates a new struct field.
func (m *structPreparer) newStructField(col string) selectItem {
	fv, ok := m.StructType.FieldByNameFunc(func(field string) bool {
		return m.matchesField2Col(field, col)
	})

	if ok {
		return &structItem{field: &fv}
	}

	return nil
}

func (m *structPreparer) matchesField2Col(field, col string) bool {
	f, _ := m.StructType.FieldByName(field)
	if v := f.Tag.Get("name"); v != "" && v != "-" {
		return v == col
	}

	eq := strings.EqualFold

	return eq(field, col) || eq(field, ss.ToCamel(col))
}

// ContainsFold tell if a contains b in case-insensitively.
func ContainsFold(a, b string) bool {
	return strings.Contains(strings.ToUpper(a), strings.ToUpper(b))
}
