package sqlmap_test

import (
	"fmt"
	"testing"

	"github.com/bingoohuang/ngg/godbtest/sqlmap/rowscan"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Log(fmt.Print(float64(1518941)))
	assert.Equal(t, "table_name123", rowscan.GetSingleTableName("select * from table_name123"))
	assert.Equal(t, "table_name123", rowscan.GetSingleTableName("select * from table_name123 abc"))
	assert.Equal(t, "table_name123", rowscan.GetSingleTableName("select * from table_name123 as abc"))
}
