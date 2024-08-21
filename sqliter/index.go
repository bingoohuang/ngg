package sqliter

import (
	"fmt"
	"strings"
)

// TableIndexInfo 表索引对象
type TableIndexInfo struct {
	// Tags 表的索引字段
	Tags []string
	// HasPk 是否有主键
	HasPk bool
}

// ParseTableIndexInfo 解析表的索引字段
func ParseTableIndexInfo(db *DebugDB, table string) (*TableIndexInfo, error) {
	q := fmt.Sprintf(`select name from sqlite_master where tbl_name = '%s' and type = 'index'`, table)

	type sqliteMaster struct {
		Name string `name:"name"`
	}

	result, err := db.Query(sqliteMaster{}, q)
	if err != nil {
		return nil, err
	}

	ti := &TableIndexInfo{}
	prefix := "idx_" + table + "_"
	rows := result.Rows.([]sqliteMaster)

	for _, row := range rows {
		if strings.HasPrefix(row.Name, prefix) {
			tag := row.Name[len(prefix):]
			ti.Tags = append(ti.Tags, tag)
		} else if strings.HasPrefix(row.Name, "udx_") {
			ti.HasPk = true
		}
	}

	return ti, nil
}
