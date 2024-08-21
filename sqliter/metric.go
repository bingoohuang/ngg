package sqliter

import (
	"log"

	"github.com/bingoohuang/ngg/sqliter/influx"
)

// WriteMetric 写入指标
// 此过程，会涉及到建库、建表/索引，或者已有库的修正表及索引
func (q *Sqliter) WriteMetric(metric influx.Metric) error {
	table := metric.Name()
	if err := q.ValidateTable(table); err != nil {
		return err
	}

	// 一表一库
	tableDB, err := q.getWriteTable(metric.Time(), table)
	if err != nil {
		return err
	}

	tableDB.makeSureTableCreated(metric, q.SeqKeysDB, q.AsTags)
	columns := createColumnsFromMetric(metric, q.SeqKeysDB)

	tableDB.insertCh <- Insert{
		metric:  metric,
		columns: columns,
	}

	return nil
}

// createColumnsFromMetric 从指标 metric 中生成列信息
func createColumnsFromMetric(metric influx.Metric, keysSeq *BoltSeq) map[string]any {
	columns := map[string]any{
		"timestamp": metric.Time(),
	}

	keysSeqFn := func(k string) any { return k }
	if keysSeq != nil {
		keysSeqFn = func(k string) any {
			seq, err := keysSeq.Next(k)
			if err != nil {
				log.Printf("E! key seq for %s: %v", k, err)
				return k
			}
			return seq
		}
	}

	for k, v := range metric.Tags() {
		columns[k] = keysSeqFn(v)
	}
	for k, v := range metric.Fields() {
		if str, ok := v.(string); ok {
			columns[k] = keysSeqFn(str)
		} else {
			columns[k] = ToDBFieldValue(v)
		}
	}
	return columns
}
