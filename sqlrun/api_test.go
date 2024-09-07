package sqlrun_test

import (
	"database/sql"
	"testing"

	"github.com/bingoohuang/ngg/sqlrun"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	libVersion, libVersionNumber, sourceID := sqlite3.Version()
	// libVersion: "3.46.1"
	// libVersionNumber: "2024-08-13 09:16:08 c9c2ab54ba1f5f46360f1b4f35d849cd3f080e6fc2b6c60e91b16c63f69a1e33"
	// sourceID: 3046001
	t.Log(libVersion, libVersionNumber, sourceID)

	d, err := sql.Open("sqlite3", ":memory:")
	assert.Nil(t, err)
	defer d.Close()

	r, err := sqlrun.Query(d, "SELECT sqlite_version() as version")
	assert.Nil(t, err)
	assert.Equal(t, 1, r.RowsCount)
	assert.Equal(t, [][]string{{libVersion}}, r.Rows.([][]string))

	type sqliteVersion struct {
		Version string
	}

	r, err = sqlrun.
		SetResultType(sqliteVersion{}).
		Query(d, "SELECT sqlite_version() as version")
	assert.Nil(t, err)
	assert.Equal(t, []sqliteVersion{{Version: libVersion}}, r.Rows.([]sqliteVersion))

	_, err = sqlrun.Query(d, "create table t2(id int primary key, name text )")
	assert.Nil(t, err)

	r, err = sqlrun.Query(d, "insert into t2(id, name) values (?, ?)", 1, "bingoo")
	assert.Nil(t, err)
	assert.Equal(t, int64(1), r.RowsAffected)

	r, err = sqlrun.Query(d, "update t2 set name = ? where id = ? ", "huang", 1)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), r.RowsAffected)

	r, err = sqlrun.Query(d, "select id, name from t2 where id = ? ", 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, r.RowsCount)
	assert.Equal(t, [][]string{{"1", "huang"}}, r.Rows.([][]string))

	type t2 struct {
		ID   int
		Name string
	}

	r, err = sqlrun.SetResultType(t2{}).Query(d, "select id, name from t2 where id = ? ", 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, r.RowsCount)
	assert.Equal(t, []t2{{ID: 1, Name: "huang"}}, r.Rows.([]t2))

	r, err = sqlrun.Query(d, "delete from t2 where id = ? ", 1)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), r.RowsAffected)
}
