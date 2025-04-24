package conf

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/sqlparser"
	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/tick"
	"github.com/cespare/xxhash/v2"
	"github.com/cheggaaa/pb/v3"
	"github.com/deatil/go-cryptobin/hash/sm3"
	cmap "github.com/orcaman/concurrent-map/v2"
	"go.uber.org/atomic"
	"golang.org/x/net/context"
)

func NewPerf(query string, options *replOptions, thinkTime *tick.ThinkTime) *Perf {
	return &Perf{
		query:   query,
		options: options,

		thinkTime: thinkTime,
	}
}

type Perf struct {
	bar Bar

	thinkTime *tick.ThinkTime
	ch        chan Batch

	options  *replOptions
	query    string
	subEvals []ss.Subs

	queryRows atomic.Int64
	affected  atomic.Int64
	errs      atomic.Int64

	currentNum atomic.Int64
	argIndex   atomic.Int64
	txBatch    int

	isSelect bool

	throttle func() bool
}

type Bar interface {
	Add(value int)
	Finish()
}

type pbBar struct {
	*pb.ProgressBar
	originNum int
	total     int64
}

func (n pbBar) Finish() { n.ProgressBar.Finish() }
func (n *pbBar) Add(value int) {
	if n.originNum == 0 {
		n.total += int64(value)
		n.ProgressBar.SetTotal(n.total)
	}
	n.ProgressBar.Add(value)
}

type noopBar struct{}

func (n noopBar) Finish() {}
func (n noopBar) Add(int) {}

type Batch struct {
	LastQuery string
	NArgs     int
}

func (f *Perf) run() error {
	if f.options.num > 0 && f.options.batch > f.options.num {
		f.options.batch = f.options.num
		f.options.threads = 1
	}

	f.txBatch = 1
	var driverName string
	if len(f.options.dss) > 0 {
		driverName = f.options.dss[0].DriverName
	}

	// https://stackoverflow.com/a/1609688/14077979
	// SQLite 3.7.11 支持多值插入
	// 例如: INSERT INTO MyTable ( Column_foo, Column_CreatedOn)
	//    VALUES ('foo 1', '2023-02-20 14:10:00.001'), ('foo 2', '2023-02-20 14:10:00.002')
	switch driverName {
	case "mysql", "pgx", "postgres", "sqlite", "sqlite3":
	default:
		f.txBatch = f.options.batch
		f.options.batch = 1
	}

	var err error
	f.query, err = ss.ExpandAtFile(f.query)
	if err != nil {
		return err
	}

	dialectFn := genDialect(driverName)
	parsedQuery, subEvals, err := ParseSQL(dialectFn, f.query)
	if err != nil {
		return err
	}

	if len(subEvals) > 0 {
		f.subEvals = subEvals
	} else {
		f.subEvals = fixBindPlaceholdersPrefix(parsedQuery, dialectFn)
	}
	fullSQL := sqlparser.String(parsedQuery) // insert ...
	var values string

	baseQuery := fullSQL
	preparedQuery := fullSQL
	printPreparedQuery := preparedQuery
	switch qt := parsedQuery.(type) {
	case *sqlparser.Insert:
		insertRowsSubSQL := sqlparser.String(qt.Rows) // values (?)
		values = strings.TrimSpace(insertRowsSubSQL[len("values"):])
		valuesPos := strings.Index(fullSQL, insertRowsSubSQL)
		baseQuery = fullSQL[:valuesPos] + " values "

		preparedQuery = baseQuery + RepeatValues(values, dialectFn, f.options.batch)
		printPreparedQuery = baseQuery + values
	case *sqlparser.Select:
		f.options.batch = 1
		f.isSelect = true
	case *sqlparser.Update, *sqlparser.Delete:
		f.txBatch = f.options.batch
		f.options.batch = 1
	default:
		if f.options.asQuery {
			f.options.batch = 1
			f.isSelect = true
		} else if f.options.asExec {
			f.txBatch = f.options.batch
			f.options.batch = 1
		} else {
			log.Printf("unknown type: %v, try %%set --asQuery / --asExec", qt)
			return err
		}
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	desc := fmt.Sprintf("threads(goroutines): %d, txBatch: %d, batch: %d", f.options.threads, f.txBatch, f.options.batch)
	if f.options.num > 0 {
		desc += fmt.Sprintf(" with %d request(s)", f.options.num)
	}
	if f.options.perfDuration > 0 {
		desc += fmt.Sprintf(" for %s", f.options.perfDuration)
		time.AfterFunc(f.options.perfDuration, cancelFunc)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, os.Kill) // handle ctrl-c

	go func() {
		for i := 0; ; i++ {
			<-sigs
			if i == 0 {
				cancelFunc()
			} else {
				os.Exit(-1)
			}
		}
	}()

	log.Printf("%s", desc)
	log.Printf("preparedQuery: %s", printPreparedQuery)

	f.throttle = func() bool { return ctx.Err() == nil }
	if f.options.qps > 0 {
		t := time.NewTicker(time.Duration(1e6/(f.options.qps)) * time.Microsecond)
		defer t.Stop()
		f.throttle = func() bool {
			select {
			case <-t.C:
				return true
			case <-ctx.Done():
				return false
			}
		}
	}

	f.ch = make(chan Batch, f.options.threads)
	go func() {
		defer close(f.ch)

		if f.options.num <= 0 { // 总数小于等于0，一直跑测试
			for {
				select {
				case <-ctx.Done():
					return
				case f.ch <- Batch{NArgs: f.options.batch}:
					f.currentNum.Add(int64(f.options.batch))
				}
			}
		}

		n := f.options.num
		for ; n >= f.options.batch; n -= f.options.batch {
			select {
			case <-ctx.Done():
				return
			case f.ch <- Batch{NArgs: f.options.batch}:
				f.currentNum.Add(int64(f.options.batch))
			}
		}

		if n > 0 {
			q1 := ss.Repeat(values, ",", n)
			select {
			case <-ctx.Done():
				return
			case f.ch <- Batch{NArgs: n, LastQuery: baseQuery + q1}:
				f.currentNum.Add(int64(n))
			}
		}
	}()

	f.executeBatch(context.Background(), preparedQuery, f.options.num)
	return nil
}

func RepeatValues(values string, fn BindNameAware, batch int) string {
	valuesExpr := values
	seq := fn.CurrentSeq()
	for i := 1; i < batch; i++ {
		subValue := values
		for j := 1; j <= seq; j++ {
			old := fn.BindName(j)
			reg := regexp.QuoteMeta(old) + `\b`
			newName := fn.BindName(j + i*seq)
			subValue = regexp.MustCompile(reg).ReplaceAllLiteralString(subValue, newName)
		}

		valuesExpr += "," + subValue
	}

	return valuesExpr
}

func (f *Perf) logErr(err error) {
	if !errors.Is(err, context.Canceled) {
		log.Printf("error occurred: %+v", err)
		f.errs.Inc()
	}
}

func (f *Perf) batchProcess(ctx context.Context, query string) error {
	tx, err := f.beginTx(ctx, f.txBatch)
	if err != nil {
		return err
	}

	for b := range f.ch {
		f.throttle()

		if b.LastQuery != "" { // 最后一个，不满足一批，强制提交
			if err := tx.commitTx(ctx, true, b.LastQuery); err != nil {
				return err
			}
		}

		if err := tx.execute(ctx, query, f.execute, b.NArgs); err != nil {
			return err
		}

		f.bar.Add(b.NArgs)

		if err := tx.commitTx(ctx, false, query); err != nil {
			return err
		}
	}

	return tx.complete()
}

func (f *Perf) executeBatch(ctx context.Context, query string, originNum int) {
	f.bar = f.createBar(originNum)
	cost := GoWait(f.options.threads, func(threadNum int) error {
		return f.batchProcess(ctx, query)
	}, f.logErr)

	f.bar.Finish()
	du := time.Duration(cost.Nanoseconds() / f.currentNum.Load())

	if f.isSelect {
		log.Printf("Average %s/record, total cost: %s, total rows: %d, errors: %d",
			du, cost, f.queryRows.Load(), f.errs.Load())
	} else {
		log.Printf("Average %s/record, total cost: %s, total affected: %d, errors: %d",
			du, cost, f.affected.Load(), f.errs.Load())
	}
}

func (f *Perf) createBar(originNum int) Bar {
	if f.options.verbose > 0 {
		return &noopBar{}
	}

	return &pbBar{ProgressBar: pb.StartNew(f.options.num).SetTemplate(pb.Full), originNum: originNum}
}

type Tx struct {
	tx    *sql.Tx
	gen   *Perf
	ps    *sql.Stmt
	count int
	batch int
}

func (f *Perf) beginTx(ctx context.Context, batch int) (*Tx, error) {
	if f.options.dryRun {
		return &Tx{gen: f, batch: batch}, nil
	}
	if ta, ok := f.options.dss[0].db.(TxContextAware); ok {
		tx, err := ta.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		return &Tx{gen: f, tx: tx, batch: batch}, nil
	}

	return nil, fmt.Errorf("db is not TxContextAware")
}

type TxContextAware interface {
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

func (f *Tx) commitTx(ctx context.Context, force bool, query string) error {
	if f.tx == nil || f.count == 0 || f.gen.options.dryRun {
		return nil
	}

	if !force && f.count < f.batch {
		f.count++
		return nil
	}

	if err := f.tx.Commit(); err != nil {
		// 提交事务，只记录失败而不返回
		// 因为可能由于并发更新时，相同记录被更新，导致提交失败
		f.gen.logErr(err)
	}

	f.count = 0
	f.tx = nil

	if query != "" {
		if ta, ok := f.gen.options.dss[0].db.(TxContextAware); ok {
			if tx, err := ta.BeginTx(ctx, nil); err != nil {
				return err
			} else {
				f.tx = tx
			}
		}

		ss.Close(f.ps)

		var err error
		if f.ps, err = f.tx.PrepareContext(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

func (f *Tx) complete() error {
	if err := f.commitTx(nil, true, ""); err != nil {
		return err
	}

	ss.Close(f.ps)

	return nil
}

func (f *Tx) execute(
	ctx context.Context,
	query string,
	execFunc func(ctx context.Context, ps *sql.Stmt, argBatch int) error,
	argBatch int,
) error {
	if err := f.tryPrepare(ctx, query); err != nil {
		return err
	}

	if err := execFunc(ctx, f.ps, argBatch); err != nil {
		return err
	}

	f.count++
	return nil
}

func (f *Tx) tryPrepare(ctx context.Context, query string) error {
	if f.ps != nil || f.gen.options.dryRun {
		return nil
	}

	ps, err := f.tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}

	f.ps = ps
	return nil
}

func (f *Perf) execute(ctx context.Context, ps *sql.Stmt, argBatch int) error {
	args, err := genArgs(argBatch, f.subEvals)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		if f.options.verbose == 1 {
			fmt.Printf("%d: %s\n", f.argIndex.Inc(), joinArgs(args, f.options.maxLen))
		} else if f.options.verbose >= 2 {
			fmt.Printf("%d: %s\t", f.argIndex.Inc(), joinArgs(args, f.options.maxLen))
		}
	}

	if f.thinkTime != nil {
		f.thinkTime.Think(true)
	}

	if f.options.dryRun {
		return nil
	}

	if f.isSelect {
		rows, err := ps.QueryContext(ctx, args...)
		if err != nil {
			return err
		}
		defer ss.Close(rows)

		for rows.Next() {
			f.queryRows.Inc()
		}

		return nil
	}

	result, err := ps.ExecContext(ctx, args...)
	if err != nil {
		return err
	}

	if affected, _ := result.RowsAffected(); affected > 0 {
		f.affected.Add(affected)
	}

	return nil
}

func joinArgs(args []any, maxLen int) string {
	if len(args) == 1 {
		return fmt.Sprintf("[%v]", ss.AbbreviateAny(args[0], maxLen, "…"))
	}

	var s string
	for i, v := range args {
		s += fmt.Sprintf("[%d: %v]", i, ss.AbbreviateAny(v, maxLen, "…"))
	}
	return s
}

func genArgs(batch int, subEvals []ss.Subs) ([]any, error) {
	substituter := NewCachingSubstituter()

	args := make([]any, 0, batch*len(subEvals))
	for i := 0; i < batch; i++ {
		for _, f := range subEvals {
			ret, err := f.Eval(substituter)
			if err != nil {
				return nil, err
			}
			args = append(args, ret)
		}
	}
	return args, nil
}

func NewCachingSubstituter() jj.Substitute {
	internal := jj.NewSubstituter(jj.DefaultSubstituteFns)
	return &cacheValuer{
		Map:      make(map[string]any),
		internal: internal,
	}
}

type cacheValuer struct {
	Map      map[string]any
	internal *jj.Substituter
}

func (v *cacheValuer) Register(fn string, f jj.SubstituteFn) {
	v.internal.Register(fn, f)
}

func (v *cacheValuer) UsageDemos() []string {
	return v.internal.UsageDemos()
}

var (
	longCache   = cmap.New[any]()
	cacheSuffix = regexp.MustCompile(`^(.+)_\d+`)
)

func (v *cacheValuer) Value(name, params, expr string) (any, error) {
	pureName := name

	subs := cacheSuffix.FindStringSubmatch(name)
	hasCachingResultTip := len(subs) > 0
	if hasCachingResultTip { // CachingSubstituter tips found
		pureName = subs[1]
		x, ok := v.Map[name]
		if !ok {
			x, ok = longCache.Get(name)
		}
		if ok {
			return x, nil
		}
	}

	if h := createHash(pureName); h != nil {
		h.Write(v.hashData(params))
		return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
	}

	x, err := v.internal.Value(pureName, params, expr)
	if err != nil {
		return nil, err
	}

	v.Map[name] = x

	if hasCachingResultTip {
		longCache.Set(name, x)
	}

	return x, nil
}

func createHash(pureName string) hash.Hash {
	switch pureName {
	case "md5":
		return md5.New()
	case "sha512":
		return sha512.New()
	case "sha256":
		return sha256.New()
	case "xxhash":
		return xxhash.New()
	case "sm3":
		return sm3.New()
	}
	return nil
}

func (v *cacheValuer) hashData(params string) []byte {
	if value, exists := v.Map[params]; exists {
		if s, ok := value.(string); ok {
			return []byte(s)
		}
		if s, ok := value.([]byte); ok {
			return s
		}
	}

	return []byte(params)
}
