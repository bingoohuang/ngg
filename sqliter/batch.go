package sqliter

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/sqliter/influx"
)

// batchInsertLoop 处理批量插入
func (d *writeTable) batchInsertLoop(c *Config) {
	tb := &tableBatch{
		writeTable:  d,
		Config:      c,
		batchVarMap: map[string]*batchVar{},
	}

	// 按批次数量/时间间隔，批量插入表数据
	if err := Tick(c.BatchInsertInterval, 3*time.Second,
		d.insertCh, tb.batchItem, tb.batchTick); err != nil {
		log.Printf("E! batch insert loop err: %+v", err)
	}

	for _, v := range tb.batchVarMap {
		if err := v.Close(); err != nil {
			log.Printf("E!  batchStmt close: %v", err)
		}
	}

	d.insertWait <- struct{}{}
}

// batchItem 将 i 加入批量处理
func (d *tableBatch) batchItem(i Insert) error {
	names, values := sortColumns(i.columns)
	query := fmt.Sprintf("insert into %q(%s) values", i.metric.Name(), strings.Join(QuoteSlice(names), ","))
	bv, ok := d.batchVarMap[query]
	if !ok {
		bv = &batchVar{
			onConflict: d.onConflictFn(names),
		}
		d.batchVarMap[query] = bv

		var err error
		if bv.Prepared, err = d.writeTable.prepareQuery(query, len(names), bv.onConflict, i.metric); err != nil {
			log.Printf("E! batch insert %s loop err: %+v", d.Table, err)
			return fmt.Errorf("prepareQuery err:%w", err)
		}
	}

	bv.addRowArgs(d, values)
	return nil
}

func (d *tableBatch) batchTick() error {
	for k, i := range d.batchVarMap {
		i.tickBatch(k, d)
	}

	d.recycle()
	return nil
}

// recycle 回收超期对象
func (d *tableBatch) recycle() {
	for k, bv := range d.batchVarMap {
		dur := time.Since(bv.lastUsed)
		if dur < d.BatchInsertInterval*10 {
			continue
		}

		log.Printf("prepared stmt %q idle for %s, try to close it", k, dur)
		if err := bv.Close(); err != nil {
			log.Printf("E!  batchStmt close:%v", err)
		}
		delete(d.batchVarMap, k)
	}
}

type Insert struct {
	metric  influx.Metric
	columns map[string]any
}

type tableBatch struct {
	*Config
	*writeTable

	batchVarMap map[string]*batchVar
}

type batchVar struct {
	// 字段名称，便于时间间隔到了，组装直接执行的 SQL
	onConflict string
	// 当前累积的绑定参数
	args []any
	// 当前已经累积的行数
	rows int

	// prepare 好的语句
	*Prepared
	// 最近使用时间，用于超期回收判断
	lastUsed time.Time
}

func (v *batchVar) tickBatch(query string, tb *tableBatch) {
	if v.rows == 0 {
		return
	}

	// 因为达不到批量次数，所以 SQL 需要根据实际的行数每次重新生成，然后直接执行
	query += MultiInsertBinds(len(v.args)/v.rows, v.rows) + v.onConflict
	r := tb.db.dbExec(false, query, v.args...)
	tb.dealError(r.Error, query, QueryExec)
	v.resetBatch()
}

func (v *batchVar) addRowArgs(d *tableBatch, values []any) {
	v.args = append(v.args, values...)
	v.rows++
	v.lastUsed = time.Now()

	// 累积行数导到批量大小，提交批量处理
	if v.rows >= v.batchSize {
		v.submitBatch(d)
	}
}

// submitBatch 提交批量插入
func (v *batchVar) submitBatch(tb *tableBatch) {
	start := time.Now()
	r, err := v.Exec(v.args...)
	cost := time.Since(start)
	if err == nil && tb.Debug {
		ra, _ := r.RowsAffected()
		log.Printf("batch insert %s rowsAffected: %d, expectRows: %d, cost: %s", tb.Table, ra, v.rows, cost)
	}
	tb.dealError(err, "batch insert "+tb.Table, BatchExec)
	v.resetBatch()
}

func (v *batchVar) resetBatch() {
	v.args = v.args[:0]
	v.rows = 0
	v.lastUsed = time.Now()
}
