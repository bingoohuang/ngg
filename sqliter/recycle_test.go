package sqliter

import (
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"

	"github.com/bingoohuang/ngg/sqliter/influx"
	"github.com/golang-module/carbon/v2"
	"github.com/stretchr/testify/assert"
)

func TestRecyle(t *testing.T) {
	f, err := os.Open("testdata/metrics.json")
	assert.Nil(t, err)
	defer f.Close()

	plus, err := New(
		WithDriverName("sqlite3"),
		WithPrefix("testdata/recycle.t"),
		WithDebug(true),
	)
	assert.Nil(t, err)

	d := json.NewDecoder(f)
	for {
		var metric influx.Point
		err := d.Decode(&metric)
		if err == io.EOF {
			break
		}
		assert.Nil(t, err)
		tt := time.Unix(metric.Timestamp, 0)
		metric.MetricTime = carbon.CreateFromStdTime(tt).SubMonth().StdTime()
		assert.Nil(t, plus.WriteMetric(&metric))
	}

	plus.tickRecycle(*plus.TimeSeriesKeep)
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
