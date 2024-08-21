package sqliter

import (
	"testing"

	"github.com/bingoohuang/ngg/sqliter/influx"
	"github.com/nxadm/tail"
	"github.com/stretchr/testify/assert"
)

func TestBatch(t *testing.T) {
	tf, err := tail.TailFile("testdata/net.log", tail.Config{})
	assert.Nil(t, err)

	plus, err := New(
		WithDriverName("sqlite3"),
		WithPrefix("testdata/batch.t"),
		//WithDebug(true),
	)
	assert.Nil(t, err)

	// Print the text of each received line
	for line := range tf.Lines {
		p, err := influx.ParseLineProtocol(line.Text)
		assert.Nil(t, err)

		err = plus.WriteMetric(p)
		assert.Nil(t, err)
	}

	assert.Nil(t, plus.Close())
	RemoveFilesPrefix(plus.SeqKeysDBName, plus.Debug)

	tables, err := plus.ListDiskTables()
	assert.Nil(t, err)

	for _, dbFiles := range tables {
		for _, dbFile := range dbFiles {
			RemoveFilesPrefix(dbFile.File.Path, plus.Debug)
		}
	}
}
