package drivers

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestSqlMode(t *testing.T) {
	db, err := sql.Open("mysql", "root:root@tcp(localhost:13306)/")
	if err != nil {
		t.Log(err)
		return
	}
	defer db.Close()

	rows, err := db.QueryContext(context.TODO(), "SELECT @@sql_mode")
	if err != nil {
		t.Log(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var result string
		err = rows.Scan(&result)
		if err != nil {
			t.Log(err)
		}
		t.Log(result)
	}
}
