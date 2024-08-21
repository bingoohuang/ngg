package sqlrun

import (
	"database/sql"
	"reflect"
)

// newStructPreparer creates a new structPreparer.
func newStructPreparer(v any) *structPreparer {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		panic(t.String() + " hasn't' not a struct or pointer to ptr type")
	}

	return &structPreparer{
		StructType: t,
	}
}

// structPreparer is the structure to create struct mapping.
type structPreparer struct {
	StructType reflect.Type
}

// Prepare prepares to scan query rows.
func (m *structPreparer) Prepare(rows *sql.Rows, columns []string) mapping {
	return &structMapping{
		rows:           rows,
		mapFields:      m.newStructFields(columns),
		structPreparer: m,
		rowsData:       reflect.MakeSlice(reflect.SliceOf(m.StructType), 0, 0),
	}
}

// structMapping is the structure for mapping row to a structure.
type structMapping struct {
	mapFields selectItemSlice
	*structPreparer
	rows     *sql.Rows
	rowsData reflect.Value
}

// Scan scans the query result to fetch the rows one by one.
func (s *structMapping) Scan(rowNum int) error {
	pointers, structPtr := s.mapFields.ResetDestinations(s.structPreparer)

	err := s.rows.Scan(pointers...)
	if err != nil {
		return err
	}

	for i, field := range s.mapFields {
		if p, ok := pointers[i].(*nullAny); ok {
			field.SetField(p.getVal())
		} else {
			field.SetField(reflect.ValueOf(pointers[i]).Elem())
		}
	}

	elem := structPtr.Elem()
	s.rowsData = reflect.Append(s.rowsData, elem)

	return nil
}

// RowsData returns the mapped rows data.
func (s *structMapping) RowsData() any { return s.rowsData.Interface() }

func (mapFields selectItemSlice) ResetDestinations(mapper *structPreparer) ([]any, reflect.Value) {
	pointers := make([]any, len(mapFields))
	structPtr := reflect.New(mapper.StructType)

	for i, fv := range mapFields {
		fv.SetRoot(structPtr.Elem())

		if implSQLScanner(fv.Type()) {
			pointers[i] = reflect.New(fv.Type()).Interface()
		} else {
			pointers[i] = &nullAny{Type: fv.Type()}
		}
	}

	return pointers, structPtr
}
