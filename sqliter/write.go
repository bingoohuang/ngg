package sqliter

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bingoohuang/ngg/sqliter/influx"
	"github.com/bingoohuang/ngg/ss"
	"github.com/elliotchance/pie/v2"
)

// getWriteTable 根据时间 t 和表名 table 获取 *writeTable 对象
func (q *Sqliter) getWriteTable(t time.Time, table string) (*writeTable, error) {
	dividedBy := q.DividedBy.DividedString(t)
	dbFile := q.TableFileBase(table, dividedBy)

	q.writeDbsLock.Lock()
	defer q.writeDbsLock.Unlock()

	if db := q.writeDbs[dbFile]; db != nil {
		// 如果错误不是需要删除文件(破坏了)，则直接返回
		if !db.LastError.shouldRemove(q.AllowDBErrors) {
			db.Last = time.Now()
			return db, nil
		}

		// 清理破坏的数据库文件
		delete(q.writeDbs, dbFile)
		db.Close()
		RemoveFilesPrefix(q.Prefix+dbFile, q.Debug)
		// 然后继续后续逻辑，重新创建库文件
	}

	dsn := q.Prefix + dbFile + "?" + q.WriteDsnOptions
	debugDB, err := NewDebugDB(dsn, 1, q.Config)
	if err != nil {
		return nil, err
	}

	wt := &writeTable{
		Config:    q.Config,
		Table:     table,
		DividedBy: dividedBy,

		db:   debugDB,
		Last: time.Now(),

		insertCh:   make(chan Insert),
		insertWait: make(chan struct{}),
	}
	go wt.batchInsertLoop(q.Config)
	q.writeDbs[dbFile] = wt
	return wt, nil
}

type writeTable struct {
	// lock 主要用于建表和改表的操作锁定（因为涉及到多条SQL执行）
	// 其它单条 SQL 执行，不需要锁定，因为写库，最大连接数已经设置为1
	lock sync.Mutex

	*Config

	db *DebugDB
	// Table 表名
	Table string
	// DividedBy 时间划分
	DividedBy string

	// Last 上次读写时间
	Last time.Time

	// tableMeta 表元信息，包括全部字段、索引字段、普通字段、是否有主键等
	tableMeta
	// LastError 记录上次操作错误，便于在下一次操作师进行识别，是否需要删除坏库文件
	LastError

	// insertCh 用于批量处理的 Channel
	insertCh chan Insert
	// insertWait 用于等待协程结束
	insertWait chan struct{}
}

type tableMeta struct {
	// Headers 表头
	// 用途1: 用于判断表是否已经创建
	Headers map[string]bool

	// HasPk 表是否有主键
	HasPk bool
	// Tags 表的索引字段
	Tags map[string]bool
	// Fields 非索引字段
	Fields map[string]bool

	// onConflictFn 生成 on conflict 子语句函数
	onConflictFn func(columns []string) string
}

type LastError struct {
	Error          error
	ErrorQuery     string
	ErrorQueryType QueryType
}

func (e *LastError) shouldRemove(allowDBErrors []string) bool {
	return e.Error != nil && !ss.ContainsAny(e.Error.Error(), allowDBErrors...)
}

func (e *LastError) dealError(err error, query string, queryType QueryType) {
	if err != nil {
		e.Error = err
		e.ErrorQuery = query
		e.ErrorQueryType = queryType
	}
}

type QueryType string

const (
	BatchPrepare QueryType = "batch-prepare"
	BatchExec    QueryType = "batch-exec"
	QueryExec    QueryType = "exec"
)

// Close 关闭数据库
func (d *writeTable) Close() {
	close(d.insertCh)
	<-d.insertWait

	d.db.Close()
}

func (d *writeTable) makeSureTableCreated(metric influx.Metric, seq *BoltSeq, tags Matcher) {
	// 涉及到检查表存在、创建表的流程，因此用一把锁保护起来
	d.lock.Lock()
	defer d.lock.Unlock()

	if len(d.Headers) > 0 {
		return
	}

	var headers []string
	r, err := d.db.Query(nil, fmt.Sprintf("select * from %q limit 1", d.Table))
	if r != nil {
		headers = r.Headers
	}

	if err != nil || r.RowsCount == 0 { // 表不存在时，创建表
		tableColumns, scripts := CreateCreateTable(metric, seq, tags)
		for _, script := range scripts {
			d.db.dbExec(true, script)
		}
		headers = tableColumns
	}

	ti, err := ParseTableIndexInfo(d.db, d.Table)
	if err != nil {
		log.Printf("failed to parse table index %s: %v", d.Table, err)
		return
	}

	d.HasPk = ti.HasPk
	d.Headers = ss.ToSet(headers)
	d.Tags = ss.ToSet(ti.Tags)
	d.Fields = ss.ToSet(pie.Filter(headers, func(s string) bool { return !d.Tags[s] }))
	d.onConflictFn = OnConflictGen(d.Tags)
}

func (d *writeTable) prepareQuery(query string, columns int, onConflict string, metric influx.Metric) (*Prepared, error) {
	// 涉及到可能的表字段/索引新增等修正，因此用一把锁保护起来
	d.lock.Lock()
	defer d.lock.Unlock()

	postfixFn := func(batchInsertSize int) string {
		return MultiInsertBinds(columns, batchInsertSize) + onConflict
	}
	stmt, err := d.db.dbPrepare(query, d.BatchInsertSize, postfixFn)
	if err != nil {
		if ss.ContainsAny(err.Error(), d.AllowDBErrors...) {
			alterTables := CreateAlterTable(&d.tableMeta, metric, d.SeqKeysDB, d.AsTags)
			for _, q := range alterTables {
				d.db.dbExec(true, q)
			}
			stmt, err = d.db.dbPrepare(query, d.BatchInsertSize, postfixFn)
		} else {
			d.dealError(err, query, BatchPrepare)
			return nil, err
		}
	}

	return stmt, err
}
