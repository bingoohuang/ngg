package sqliter

import (
	"runtime"
	"time"

	"github.com/bingoohuang/ngg/sqlrun"
)

// Read 执行查询
// table 表名称
// query 查询 SQL
// dividedTime 查询落在的时间划分（哪个时间分区库上）
// emptyStruct 从结果集映射到哪个结构体上
func (q *Sqliter) Read(table, query string, dividedTime time.Time, emptyStruct any, args ...any) (*sqlrun.Result, error) {
	if err := q.ValidateTable(table); err != nil {
		return nil, err
	}

	dividedBy := q.DividedString(dividedTime)
	db, err := q.getReadDB(table, dividedBy)
	if err != nil {
		return nil, err
	}

	r, err := db.db.Query(emptyStruct, query, args...)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (q *Sqliter) getReadDB(table, dividedBy string) (*readTable, error) {
	q.readDbsLock.Lock()
	defer q.readDbsLock.Unlock()

	dbFile := q.TableFileBase(table, dividedBy)
	if db := q.readDbs[dbFile]; db != nil {
		// 如果错误不是需要删除文件(破坏了)，则直接返回
		if !db.LastError.shouldRemove(q.AllowDBErrors) {
			db.Last = time.Now()
			return db, nil
		}

		// 清理破坏的数据库文件
		delete(q.writeDbs, dbFile)
		db.Close()
		RemoveFilesPrefix(q.Prefix+dbFile, q.Debug)
		// 直接返回库不存在
		return nil, ErrNotFound
	}

	dsn := q.Prefix + dbFile + "?" + q.ReadDsnOptions
	debugDB, err := NewDebugDB(dsn, max(4, runtime.NumCPU()), q.Config)
	if err != nil {
		return nil, err
	}

	return &readTable{
		db: debugDB,
	}, nil
}

type readTable struct {
	db *DebugDB

	// Last 上次读写时间
	Last time.Time

	// LastError 记录上次操作错误，便于在下一次操作师进行识别，是否需要删除坏库文件
	LastError
}

func (d readTable) Close() {
	d.db.Close()
}
