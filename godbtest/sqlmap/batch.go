package sqlmap

import (
	"database/sql"
	"fmt"
	"time"
)

type BatchNotifierFn func(b *BatchUpdate, start, batchStart time.Time, totalNum, batchNum int, complete bool)

type BatchUpdate struct {
	DB       *sql.DB    // 数据库
	varsCh   chan []any // 参数通道
	errCh    chan error
	Notifier BatchNotifierFn
	Query    string // update/delete/inert 等需要批量执行的语句
	BatchNum int    // 一次事务当做一批，一批的数量
}

func WithBatchNotifier(notifier BatchNotifierFn) func(*BatchUpdate) {
	return func(c *BatchUpdate) { c.Notifier = notifier }
}

func WithBatchNum(batchNum int) func(*BatchUpdate) {
	return func(c *BatchUpdate) { c.BatchNum = batchNum }
}

func NewBatchUpdate(db *sql.DB, query string, fns ...func(*BatchUpdate)) *BatchUpdate {
	b := &BatchUpdate{
		DB:       db,
		BatchNum: 1000,
		Query:    query,
		varsCh:   make(chan []any),
		errCh:    make(chan error),
	}

	for _, f := range fns {
		f(b)
	}

	go func() {
		b.errCh <- b.run()
	}()
	return b
}

func (b *BatchUpdate) Close() error {
	close(b.varsCh)
	return <-b.errCh
}

func (b *BatchUpdate) AddVars(vars []any) {
	b.varsCh <- vars
}

func (b *BatchUpdate) run() error {
	var (
		err      error
		tx       *sql.Tx   // 当前事务
		stmt     *sql.Stmt // 当前语句
		totalNum int       // 当前批次累积的
		batchNum int       // 当前批次累积的
	)
	start := time.Now()
	batchStart := time.Now()

	for vars := range b.varsCh {
		if tx == nil {
			tx, err = b.DB.Begin()
			if err != nil {
				return fmt.Errorf("begin: %w", err)
			}
			if stmt, err = tx.Prepare(b.Query); err != nil {
				return fmt.Errorf("prepare query: %w", err)
			}
		}
		if _, err := stmt.Exec(vars...); err != nil {
			return fmt.Errorf("stmt Exec: %w", err)
		}

		totalNum++
		if batchNum++; batchNum == b.BatchNum {
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("tx Commit: %w", err)
			}
			tx, err = b.DB.Begin()
			if err != nil {
				return fmt.Errorf("begin: %w", err)
			}
			if stmt, err = tx.Prepare(b.Query); err != nil {
				return fmt.Errorf("prepare sql: %w", err)
			}

			if b.Notifier != nil {
				b.Notifier(b, start, batchStart, totalNum, batchNum, false)
			}

			batchNum = 0
			batchStart = time.Now()
		}
	}

	if batchNum > 0 {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("tx Commit: %w", err)
		}
	}
	if b.Notifier != nil {
		b.Notifier(b, start, batchStart, totalNum, batchNum, true)
	}

	return nil
}
