package sqliter

import (
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"

	"github.com/bingoohuang/ngg/sqliter/influx"
	"github.com/elliotchance/pie/v2"
	"github.com/golang-module/carbon/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestWriteMetricFile(t *testing.T) {
	// 测试数据中的所有点，都是2024-08-10这天生成的
	tim := carbon.CreateFromDate(2024, 8, 10).StdTime()
	WriteMetricFileDividedBy(t, tim, DividedByMonth)
	WriteMetricFileDividedBy(t, tim, DividedByWeek)
	WriteMetricFileDividedBy(t, tim, DividedByDay)
}

func WriteMetricFileDividedBy(t *testing.T, tim time.Time, dividedBy DividedBy) {
	f, err := os.Open("testdata/metrics.json")
	assert.Nil(t, err)
	defer f.Close()

	plus, err := New(
		WithDriverName("sqlite3"),
		WithPrefix("testdata/metric.t"),
		// WithDebug(true),
		WithDividedBy(dividedBy),
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
		metric.MetricTime = time.Unix(metric.Timestamp, 0)
		assert.Nil(t, plus.WriteMetric(&metric))
	}

	assert.Nil(t, plus.Close())

	tables, err := plus.ListDiskTables()
	assert.Nil(t, err)

	dbFiles := func(dbFiles map[string][]*DbFile) map[string][]string {
		result := map[string][]string{}
		for k, v := range tables {
			result[k] = pie.Map(v, func(t *DbFile) string { return t.DividedBy })
		}
		return result
	}(tables)

	dividedStr := dividedBy.DividedString(tim)
	assert.Equal(t,
		map[string][]string{
			"cpu":       {dividedStr},
			"disk":      {dividedStr},
			"diskio":    {dividedStr},
			"mem":       {dividedStr},
			"net":       {dividedStr},
			"processes": {dividedStr},
			"system":    {dividedStr},
		}, dbFiles)

	for table, dbFileMonth := range dbFiles {
		for _, month := range dbFileMonth {
			filePath := plus.TableFilePath(table, month)
			RemoveFilesPrefix(filePath, plus.Debug)
		}
	}

	RemoveFilesPrefix(plus.SeqKeysDBName, plus.Debug)
}
