package sqliter

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/tick"
	"github.com/golang-module/carbon/v2"
	"github.com/robfig/cron/v3"
	"go.uber.org/multierr"
)

// recycleLoop 回收循环，以指定的间隔循环
func (q *Sqliter) recycleLoop() {
	ctx, cancel := context.WithCancel(context.Background())
	q.recycleCancel = cancel

	if q.RecycleCron != "" {
		parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)

		schedule, err := parser.Parse(q.RecycleCron)
		if err != nil {
			log.Fatalf("fail to parse %q: %v", q.RecycleCron, err)
		}

		c := cron.New(cron.WithParser(parser))
		c.Start()
		defer c.Stop()

		// @midnight
		// @every 5m
		// 每秒: * * * * * ?
		// 每5分钟: 0 5 * * * *", every5min(time.Local)},
		c.Schedule(schedule, cron.FuncJob(func() {
			q.tickRecycle(*q.TimeSeriesKeep)
		}))

		// 等待结束
		<-ctx.Done()
	} else {
		t := time.NewTicker(q.RecycleInterval)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				q.tickRecycle(*q.TimeSeriesKeep)
			case <-ctx.Done():
				return
			}
		}
	}
}

type RecycleResult struct {
	BeforeTime      time.Time `json:"beforeTime"`
	RecycledRecords int64     `json:"recycledRecords"`
	DirSize         int64     `json:"dirSize"`
	Error           string    `json:"error"`
}

// Recycle 手动触发回收
// before 设置为空时，按照系统配置的策略执行一次回收,
// 格式1(绝对时间): RFC3339 "2006-01-02T15:04:05Z07:00"
// 格式2(偏移间隔): -10d 10天前的此时
func (q *Sqliter) Recycle(before string) (rr RecycleResult) {
	var err error
	timeSpan := *q.TimeSeriesKeep
	if before != "" {
		rr.BeforeTime, err = tick.ParseTime(before)
		if err != nil {
			rr.Error = err.Error()
			return rr
		}

		days := carbon.Now().DiffInDays(carbon.CreateFromStdTime(rr.BeforeTime))
		if days < 0 {
			rr.Error = fmt.Sprintf("bad before %s", before)
			return rr
		}
		timeSpan = TimeSpan{Value: int(days), Unit: UnitDay}
	}

	rr.DirSize = q.tickRecycle(timeSpan)
	return rr
}

// tickRecycle 按 keepTime 进行回收操作
func (q *Sqliter) tickRecycle(keepTime TimeSpan) (totalRecycledSize int64) {
	t := time.Now()
	// 回收读库，关闭过期读库(超过空闲期）
	q.tickRecycleReadDbs()
	// 回收写库，关闭过期写库（非当月，或者超过空闲期）
	q.tickRecycleWriteDbs(q.DividedString(t))

	// 按保留天数从 t 往前 keepTime 进行回收
	totalRecycledSize = q.tickRecycleByDays(t, keepTime)
	if q.TimeSeriesMaxSize > 0 {
		// 按最大保留大小从 t 往前回收
		totalRecycledSize += q.tickRecycleByMaxSize(t)
	}

	return totalRecycledSize
}

// tickRecycleByMaxSize 按最大保留大小从 t 往前回收
func (q *Sqliter) tickRecycleByMaxSize(t time.Time) (totalRecycledSize int64) {
	tables, err := q.ListDiskTables()
	if err != nil {
		log.Printf("sqliter.ListDiskTables() error: %v", err)
		return
	}

	totalSize := int64(0)
	for _, infos := range tables {
		for _, info := range infos {
			totalSize += info.TotalSize()
		}
	}

	// 总大小低于设定大小，不用回收，结束
	if totalSize <= q.TimeSeriesMaxSize {
		return
	}

	recycled := q.findRecycleFilesBySize(t, tables, totalSize)
	if len(recycled) > 0 {
		totalRecycledSize = q.tickRecycleDbFiles(recycled)
	}
	return totalRecycledSize
}

func (q *Sqliter) findRecycleFilesBySize(t time.Time, tables map[string][]*DbFile, totalSize int64) []*DbFile {
	currenDividedBy := q.DividedString(t)

	var recycled []*DbFile
	for _, infos := range tables {
		for _, info := range infos {
			if info.DividedBy >= currenDividedBy {
				continue
			}

			recycled = append(recycled, info)
			totalSize -= info.TotalSize()
			if totalSize <= q.TimeSeriesMaxSize {
				return recycled
			}
		}
	}

	return recycled
}

// tickRecycleByDays 按保留天数从 t 往前 keepTime 进行回收
func (q *Sqliter) tickRecycleByDays(t time.Time, keepTime TimeSpan) (totalRecycledSize int64) {
	tables, err := q.ListDiskTables()
	if err != nil {
		log.Printf("sqliter.ListDiskTables() error: %v", err)
		return
	}

	cutoffDivided := q.DividedBy.CutoffDays(t, keepTime)
	var recycled []*DbFile
	for _, infos := range tables {
		for _, info := range infos {
			if info.DividedBy < cutoffDivided {
				recycled = append(recycled, info)
			} else {
				log.Printf("file: %s DividedBy: %s >= cutoffDivided: %s", info.File.Path, info.DividedBy, cutoffDivided)
			}
		}
	}

	if len(recycled) > 0 {
		totalRecycledSize = q.tickRecycleDbFiles(recycled)
	}

	return
}

// tickRecycleDbFiles 按文件进行删除回收
func (q *Sqliter) tickRecycleDbFiles(recycled []*DbFile) (totalRecycledSize int64) {
	q.recycleReadDbs(recycled)
	q.recycleWriteDbs(recycled)

	for _, r := range recycled {
		_, removedSize := RemoveFilesPrefix(r.File.Path, true)
		totalRecycledSize += removedSize
	}

	log.Printf("recycleFiles: %d, size: %s", len(recycled),
		ss.IBytes(uint64(totalRecycledSize)))

	return
}

func (q *Sqliter) recycleReadDbs(recycled []*DbFile) {
	q.readDbsLock.Lock()
	defer q.readDbsLock.Unlock()

	for _, r := range recycled {
		dbFile := q.TableFileBase(r.Table, r.DividedBy)
		if db, ok := q.readDbs[dbFile]; ok {
			db.Close()
			delete(q.readDbs, dbFile)
		}
	}
}

func (q *Sqliter) recycleWriteDbs(recycled []*DbFile) {
	q.writeDbsLock.Lock()
	defer q.writeDbsLock.Unlock()

	for _, r := range recycled {
		dbFile := q.TableFileBase(r.Table, r.DividedBy)
		if db, ok := q.writeDbs[dbFile]; ok {
			db.Close()
			delete(q.writeDbs, dbFile)
		}
	}
}

// tickRecycleReadDbs 回收读库，关闭过期读库(超过空闲期）
func (q *Sqliter) tickRecycleReadDbs() {
	q.readDbsLock.Lock()
	defer q.readDbsLock.Unlock()

	for dbFile, db := range q.readDbs {
		if time.Since(db.Last) > q.MaxIdle {
			db.Close()
			delete(q.readDbs, dbFile)
		}
	}
}

// tickRecycleWriteDbs 回收写库，关闭过期写库（早于当前时间划分，或者超过空闲期）
func (q *Sqliter) tickRecycleWriteDbs(dividedBy string) {
	q.writeDbsLock.Lock()
	defer q.writeDbsLock.Unlock()

	for dbFile, db := range q.writeDbs {
		// 早于当前时间划分，或者超过空闲期
		if db.DividedBy < dividedBy || time.Since(db.Last) > q.MaxIdle {
			db.Close()
			delete(q.writeDbs, dbFile)
		}
	}
}

// DbStat 库状态统计
type DbStat struct {
	// DSN 数据源名字
	DSN string `json:"dsn"`
	// LastVisit 最后访问时间
	LastVisit time.Time `json:"lastVisit"`
	// DividedBy 时间划分字符串
	DividedBy string `json:"dividedBy"`
	// ReadOnly 是否只读
	ReadOnly bool `json:"readOnly"`
}

// StatDbs 统计读写库
func (q *Sqliter) StatDbs() (dbStats []DbStat) {
	dbStats = append(dbStats, q.statReadDbs()...)
	dbStats = append(dbStats, q.statWriteDbs()...)
	return
}

// statReadDbs 统计读库
func (q *Sqliter) statReadDbs() (dbStats []DbStat) {
	q.readDbsLock.Lock()
	defer q.readDbsLock.Unlock()

	for _, db := range q.readDbs {
		dbStats = append(dbStats, DbStat{
			DSN:       db.db.DSN,
			LastVisit: db.Last,
			DividedBy: db.DividedBy,
			ReadOnly:  true,
		})
	}
	return
}

// statWriteDbs 统计写库
func (q *Sqliter) statWriteDbs() (dbStats []DbStat) {
	q.writeDbsLock.Lock()
	defer q.writeDbsLock.Unlock()

	for _, db := range q.writeDbs {
		dbStats = append(dbStats, DbStat{
			DSN:       db.db.DSN,
			LastVisit: db.Last,
			DividedBy: db.DividedBy,
		})
	}
	return
}

// Close 关闭 sqliter 所有操作，包括关闭库文件、退出回收协程等
func (q *Sqliter) Close() error {
	q.recycleCancel()
	q.closeWriteDbs()
	q.closeReadDbs()

	var err error
	if q.SeqKeysDB != nil {
		err = multierr.Append(err, q.SeqKeysDB.Close())
	}

	return err
}

func (q *Sqliter) closeReadDbs() {
	// 锁保护 q.readDbs 遍历
	q.readDbsLock.Lock()
	defer q.readDbsLock.Unlock()

	for _, db := range q.readDbs {
		db.Close()
	}
}

func (q *Sqliter) closeWriteDbs() {
	// 锁保护 q.writeDbs 遍历
	q.writeDbsLock.Lock()
	defer q.writeDbsLock.Unlock()

	for _, db := range q.writeDbs {
		db.Close()
	}
}
