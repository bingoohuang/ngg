package conf

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/bingoohuang/ngg/godbtest/files"
	"github.com/bingoohuang/ngg/godbtest/label"
	"github.com/bingoohuang/ngg/godbtest/sqlmap"
	"github.com/bingoohuang/ngg/godbtest/ui"
	"github.com/bingoohuang/ngg/ss"
	"github.com/bingoohuang/ngg/tick"
	"github.com/creasty/defaults"
	"github.com/mattn/go-shellwords"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Prompt      string        `yaml:"prompt" default:"always"`
	ListenAddr  string        `yaml:"listenAddr"`
	ContextPath string        `yaml:"contextPath" default:"/"`
	DataSources []*DataSource `yaml:"dataSources"`

	Actions []*Action `yaml:"actions"`
	Offset  int       `yaml:"offset"`
	Limit   int       `yaml:"limit" default:"1000"`
}

func ParseConfigFile(filePath, driverName string) (*Config, error) {
	var c Config
	defer c.setDefaults()

	if filePath == "" || !ss.Pick1(ss.Exists(filePath)) {
		return &c, nil
	}

	cf, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s failed: %v", filePath, err)
	}

	if err := yaml.Unmarshal(cf, &c); err != nil {
		return nil, fmt.Errorf("unmarshal yaml file %s failed: %w", cf, err)
	}

	if driverName != "" {
		c.chooseDataSource(driverName)
	}

	return &c, nil
}

func (c *Config) Go(dsn string, evaluates, args []string) {
	var wg sync.WaitGroup
	defer wg.Wait()

	if c.ListenAddr != "" {
		wg.Add(1)
		go c.listen(&wg)
	}

	options := &replOptions{
		times:   1,
		offset:  c.Offset,
		limit:   c.Limit,
		lookups: map[string]map[string]string{},
	}

	// 初始化 perf 相关参数
	_ = parseSetArgs("", options, nil, "")

	if len(options.dss) == 0 && !options.dryRun {
		options.dss = c.getDSS(dsn, len(evaluates), options, args)
	}
	defer ss.Close(options.dss...)

	if len(evaluates) > 0 {
		options.printSQL = true
		for _, q := range evaluates {
			executeInput(q, options)
		}
		return
	}

	namedSqls := c.getNamedSqls(options.dss, options)

	for {
		if options.named && len(namedSqls) > 0 {
			i := ui.Select("Select Sql to execute", namedSqls)
			sq := namedSqls[i]
			if sq.DS == nil {
				break
			}

			sq.Run(sq.DS, options)
		}

		switch c.Prompt {
		case "always", "once":
			q, err := ui.GetSQL(options.sep)
			if err != nil {
				if errors.Is(err, ui.ErrExit) {
					return
				}

				log.Printf("error: %v", err)
				continue
			}
			executeInput(q, options)
		}

		if c.Prompt != "always" {
			break
		}
	}
}

type DSNItem struct {
	DSN   string
	Title string
}

func (d DSNItem) ItemTitle() string { return d.Title }
func (d DSNItem) ItemDesc() string  { return d.DSN }

func (c *Config) getDSS(dsn string, evaluateNum int, options *replOptions, args []string) []*DataSource {
	if dsn != "" {
		if ss.Pick1(ss.Exists(".env")) {
			envMap, _ := ReadEnvFile(".env")
			if dsns := envMap["DSN"]; len(dsns) > 1 {
				var dsnItems []DSNItem
				for _, dsn := range dsns {
					if driverName, _ := AutoTestConnect(dsn); driverName != "" {
						dsnItems = append(dsnItems, DSNItem{DSN: dsn, Title: driverName})
					}
				}

				out := ui.Select("Choose DataSource:", dsnItems)
				dsn = dsnItems[out].DSN
			}
		}

		if driverName, dataSourceName := AutoTestConnect(dsn); driverName != "" {
			return connectDB(driverName, dataSourceName, nil, options.verbose, options.maxOpenConns)
		}
	}

	if evaluateNum > 0 {
		return nil
	}

	if len(args) > 0 && ss.Pick1(ss.Exists(args[0])) {
		return connectDB("sqlite", args[0], nil, options.verbose, options.maxOpenConns)
	}

	dss := getDataSources(c.DataSources, nil)
	if len(dss) > 0 {
		out := ui.Select("Choose DataSource:", append(dss, &DataSource{
			Name: "customized",
		}))
		if out < len(dss) {
			dss = []*DataSource{dss[out]}
		} else {
			dss = dss[:0]
		}
	}

	return dss
}

func (c *Config) getNamedSqls(dss []*DataSource, options *replOptions) []DsSql {
	var namedSqls []DsSql
	for _, action := range c.Actions {
		if named := action.Go(dss, options); len(named) > 0 {
			namedSqls = append(namedSqls, named...)
		}
	}
	if len(namedSqls) > 0 {
		namedSqls = append(namedSqls, DsSql{Sql: &Sql{Name: "Custom", Sql: "Input"}})
	}
	return namedSqls
}

func executeInput(q string, options *replOptions) {
	var qq []string
	if strings.HasPrefix(q, "%") {
		qq = []string{q}
	} else if strings.HasPrefix(q, "@") {
		fileName, _ := parseSuffix(q[1:], DisplayDefault, options.sep)
		dealScriptFile(fileName, options)
		return
	} else if ss.Pick1(ss.Exists(q)) {
		dealScriptFile(q, options)
		return
	} else if subVars := parseVars(q); subVars != nil {
		qq = subVars.evalSQL(options.sep)
	} else {
		qq = ss.SplitReg(q, ui.SepReg(options.sep), -1)
	}

	executeQQ(q, options, qq)
}

func executeQQ(q string, options *replOptions, qq []string) {
	think, err := tick.ParseThinkTime(options.thinkTime)
	if err != nil {
		log.Printf("bad think time: %s", options.thinkTime)
		return
	}

	var remoteQueries []string
	for _, query := range qq {
		query = strings.TrimSpace(query)

		if strings.HasPrefix(query, `%`) {
			pureQ, _ := parseSuffix(query, DisplayDefault, options.sep)
			parser := shellwords.NewParser()
			fq, err := parser.Parse(pureQ)
			if err != nil {
				log.Printf("parse query %s failed: %v", query, err)
				return
			}

			name := fq[0]
			if parser.Position < 0 {
				pureQ = ""
			} else {
				pureQ = strings.TrimSpace(pureQ[len(name):])
			}

			if optionFunc := findOptionFunc(name); optionFunc != nil {
				if args := fq[1:]; len(args) == 0 {
					if optionFunc.getter != nil {
						optionFunc.getter(optionFunc.name, options)
					}
				} else if optionFunc.setter != nil {
					if err := optionFunc.setter(optionFunc.name, options, args, pureQ); err != nil {
						log.Printf("error: %v", err)
					}
				}
				continue
			}
		}

		if strings.HasSuffix(query, `\P`) {
			if len(options.dss) == 0 && !options.dryRun {
				log.Printf("no data sources specified, please type %%help to get help")
				continue
			}

			query = strings.TrimSuffix(query, `\P`)

			perf := NewPerf(query, options, think)
			if err := perf.run(); err != nil {
				fmt.Println(err)
			}
			continue
		}

		remoteQueries = append(remoteQueries, query)
	}

	if len(remoteQueries) == 0 {
		return
	}

	if len(options.dss) == 0 && !options.dryRun {
		log.Printf("no data sources specified, please type %%help to get help")
		return
	}

	sq := Sql{Sql: q, IgnoreError: true, qq: remoteQueries, Format: options.format}

	for i := 0; i < options.times; i++ {
		if i > 0 {
			if think != nil {
				think.Think(true)
			}
		}
		if options.dryRun {
			sq.Run(nil, options)
		} else {
			for _, d := range options.dss {
				sq.Run(d, options)
			}
		}
	}
}

func dealScriptFile(fileName string, options *replOptions) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("read file %s failed: %v", fileName, err)
		return
	}
	defer file.Close()
	start := time.Now()
	defer func() {
		fmt.Printf(">> execute %s finished in %s\n", fileName, time.Since(start))
	}()

	scanner := files.NewLineScanner(file)
	for !scanner.Stopped {
		querys, err := scanner.Scan(options.sep)
		for _, query := range querys {
			executeQQ(query, options, []string{query})
		}
		if err != nil {
			log.Printf("read file %s failed: %v", fileName, err)
			return
		}
	}
}

func connectDB(driverName, dataSourceName string, dss []*DataSource, verbose, maxOpenConns int) []*DataSource {
	ds := &DataSource{
		DriverName:     driverName,
		DataSourceName: dataSourceName,
	}
	if err := ds.Connect(false, verbose, maxOpenConns); err != nil {
		log.Printf("Connect error: %v", err)
		if strings.Contains(err.Error(), "unknown driver") {
			log.Printf("current imported drivers: %v", sql.Drivers())
		}

		return dss
	}

	ss.Close(dss...)
	return []*DataSource{ds}
}

func (c *Config) setDefaults() {
	c.Limit = lo.Ternary(c.Limit <= 0, 100, c.Limit)

	for _, ds := range c.DataSources {
		if ds.Labels == nil {
			ds.Labels = map[string]string{"driverName": ds.DriverName}
		} else {
			ds.Labels["driverName"] = ds.DriverName
		}
	}

	if err := defaults.Set(c); err != nil {
		log.Printf("defaults.Set failed: %v", err)
	}

	c.parserVars()
}

func (c *Config) chooseDataSource(driverName string) {
	lower := strings.ToLower(driverName)
	for _, ds := range c.DataSources {
		d := strings.ToLower(ds.DriverName)
		ds.Disabled = !(strings.Contains(d, lower) || ss.Pick1(ss.FnMatch(lower, d, true)))
	}
}

func getDataSources(dss []*DataSource, expr *label.Visitor) (dbs []*DataSource) {
	for _, ds := range dss {
		if ds.Disabled {
			continue
		}

		if expr != nil {
			ok, err := expr.Eval(ds.Labels)
			if err != nil {
				log.Printf("eval failed: %v", err)
			}
			if !ok {
				continue
			}
		}

		dbs = append(dbs, ds)
	}

	return dbs
}

type DataSource struct {
	db     DB
	Labels map[string]string `yaml:"labels"`

	Name            string `yaml:"name"`
	DriverName      string `yaml:"driverName"`
	DataSourceName  string `yaml:"dataSourceName"`
	DataSourceEnv   string `yaml:"dataSourceEnv"`
	currentDatabase string
	Disabled        bool `yaml:"disabled"`
}

func (d *DataSource) ItemTitle() string { return ss.Or(d.Name, d.DriverName) }
func (d *DataSource) ItemDesc() string  { return d.GetDataSourceName() }
func (d *DataSource) GetDataSourceName() string {
	return ss.Or(d.DataSourceName, os.Getenv(d.DataSourceEnv))
}

func (d *DataSource) Connect(panicOnError bool, verbose int, maxOpenConns int) (err error) {
	if d.db == nil {
		d.db, err = sqlmap.Connect(d.DriverName, d.GetDataSourceName(),
			sqlmap.WithVerbose(verbose > 0),
			sqlmap.WithMaxOpenConns(maxOpenConns),
		)
	}

	if err != nil && panicOnError {
		log.Fatal(err)
	}

	return
}

func (d *DataSource) Close() error {
	if d.db == nil {
		return nil
	}

	var err error

	if closer, ok := d.db.(io.Closer); ok {
		err = closer.Close()
	}

	d.db = nil
	return err
}

var useDbReq = regexp.MustCompile(`(?i)use\s+\S+`)

func IsUsingDB(q string) bool {
	return useDbReq.MatchString(q)
}

// TODO: use db 在 *sql.DB 连接池中，可能会失效，需要调整实现
// https://github.com/go-sql-driver/mysql/issues/173
func (d *DataSource) useDB(ctx context.Context, q string) bool {
	changeToDB := strings.TrimSpace(q[3:])
	d.currentDatabase = changeToDB
	if err := RunSQL(ctx, d.db, d.DriverName, q, RunSQLOption{Format: "none", Limit: 1}); err != nil {
		log.Printf("Run sql %s failed: %+v", q, err)
	}

	return true
}

type TxAware interface {
	Begin() (*sql.Tx, error)
}

func (d *DataSource) processTx(db DB, q string) (DB, bool) {
	switch lq := strings.ToLower(q); lq {
	case "begin":
		if txa, ok := d.db.(TxAware); ok {
			tx, err := txa.Begin()
			if err != nil {
				log.Fatalf("tx begin failed: %+v", err)
			}
			return tx, true
		}
	case "commit", "rollback":
		if tx, ok := db.(*sql.Tx); ok {
			if err := lo.Ternary(lq == "commit", tx.Commit, tx.Rollback)(); err != nil {
				log.Fatalf("tx %s failed: %+v", lq, err)
			}
			return d.db, true
		}
	}

	return db, false
}

type Action struct {
	labelExpr  *label.Visitor
	LabelQuery string `yaml:"labelQuery"`
	Sqls       []*Sql `yaml:"sqls"`

	Disabled bool `yaml:"disabled"`
}

type DsSql struct {
	DS *DataSource
	*Sql
}

func (a *Action) Go(dss []*DataSource, r *replOptions) (namedSqls []DsSql) {
	if a.Disabled {
		return
	}

	var err error

	if a.labelExpr == nil {
		a.labelExpr, err = label.Parse(a.LabelQuery)
		if err != nil {
			log.Fatalf("parse label %s failed: %v", a.LabelQuery, err)
		}
	}

	dss = getDataSources(dss, a.labelExpr)
	for _, ds := range dss {
		for _, q := range a.Sqls {
			if q.Name == "" || q.AutoExecute {
				q.Run(ds, r)
			}
			if q.Name != "" {
				namedSqls = append(namedSqls, DsSql{DS: ds, Sql: q})
			}
		}
	}

	return namedSqls
}

type Sql struct {
	RowsScanner sqlmap.RowsScanner `yaml:"-"`

	subVars *SubVars

	Name string `yaml:"name"` // SQL 名字，用于提示列表
	Sql  string `yaml:"sql"`

	// Format 格式化输出
	// NONE 不输出
	// JSON 使用 JSON 格式化输出
	// TABLE 使用表格形式的格式化
	// MARKDOWN 使用 MARKDOWN 格式化输出
	Format string `yaml:"format"`

	qq []string

	// Iterations 表示 SQL 执行次数
	Iterations int `yaml:"iterations"`

	currentRunTimes int
	AutoExecute     bool `yaml:"autoExecute"` // 有名 SQL 是否自动执行，只有在 Name 有值时起作用
	Disabled        bool `yaml:"disabled"`
	IgnoreError     bool `yaml:"ignoreError"`

	// SingleOne 只执行第一条 SQL
	SingleOne bool `yaml:"singleOne"`
}

func (c *Sql) ItemTitle() string { return c.Name }
func (c *Sql) ItemDesc() string  { return c.Sql }

func (c *Sql) Run(ds *DataSource, options *replOptions) {
	if c.Disabled {
		return
	}

	csql := c.Sql
	if len(c.qq) == 0 {
		c.qq = ss.SplitReg(csql, ui.SepReg(options.sep), -1)
	}

	iterations := ss.If(c.Iterations <= 0 || c.SingleOne, 1, c.Iterations)
	qq := c.qq

	sqlEvaluated := c.subVars != nil
	if sqlEvaluated {
		qq = c.subVars.evalSQL(options.sep)
	}

	var db DB
	if !options.dryRun {
		_ = ds.Connect(true, options.verbose, options.maxOpenConns)
		db = ds.db
	}

	ctx, cancel := WithSignal(context.Background(), options.timeout)
	defer cancel()

	for i := 0; i < iterations; i++ {
		if ctx.Err() != nil {
			return
		}

		for _, q := range qq {
			if q == "" || strings.HasPrefix(q, "--") {
				continue
			}

			printSQL := (sqlEvaluated || options.printSQL) && c.currentRunTimes == 0
			if IsUsingDB(q) {
				if printSQL {
					log.Printf("RunSQL: %s", sqlmap.Color(q))
				}
				ds.useDB(ctx, q)
				continue
			}

			if ctx.Err() != nil {
				return
			}

			q, displayMode := parseSuffix(q, DisplayDefault|DisplayVertical|DisplayInsertSQL|DisplayJSON, options.sep)
			if !options.dryRun {
				var isTx bool
				if db, isTx = ds.processTx(db, q); isTx {
					continue
				}
			}

			o := RunSQLOption{
				Format:        ss.Or(options.format, c.Format),
				Offset:        options.offset,
				Limit:         options.limit,
				DisplayMode:   displayMode,
				RowsScanner:   c.RowsScanner,
				PrintCost:     options.showCost,
				SaveResult:    options.saveResult,
				TempResultDB:  options.tempResultDB,
				Timeout:       options.timeout,
				ShowRowIndex:  options.ShowRowIndex,
				NoEvalSQL:     options.NoEvalSQL,
				RawFileDir:    options.rawFileDir,
				RawFileExt:    options.rawFileExt,
				DryRun:        options.dryRun,
				AsQuery:       options.asQuery,
				AsExec:        options.asExec,
				ParsePrepared: !options.noParsePrepared,
				MaxLen:        options.maxLen,
			}
			if options.lookupsOn {
				o.Lookup = options.lookups
			}

			if printSQL {
				log.Printf("RunSQL: %s", sqlmap.Color(q))
			}

			var driverName string
			if ds != nil {
				driverName = ds.DriverName
			}
			if err := RunSQL(ctx, db, driverName, q, o); err != nil {
				lo.Ternary(c.IgnoreError, log.Printf, log.Fatalf)("Run sql failed: %+v", err)
			}

			if c.SingleOne {
				return
			}
		}

		c.currentRunTimes++
	}
}

type DisplayMode int

const (
	DisplayDefault = 1 << iota // 1 << 0 which is 00000001
	DisplayVertical
	DisplayInsertSQL
	DisplayJSON
	DisplayJSONFree
)

func parseSuffix(q string, allowModes int, sep string) (query string, mode int) {
	switch {
	case allowModes&DisplayInsertSQL == DisplayInsertSQL && strings.HasSuffix(q, `\I`):
		return q[:len(q)-2], DisplayInsertSQL
	case allowModes&DisplayJSON == DisplayJSON && strings.HasSuffix(q, `\J`):
		return q[:len(q)-2], DisplayJSONFree
	case allowModes&DisplayJSON == DisplayJSON && strings.HasSuffix(q, `\j`):
		return q[:len(q)-2], DisplayJSON
	case allowModes&DisplayVertical == DisplayVertical && strings.HasSuffix(q, `\G`):
		return q[:len(q)-2], DisplayVertical
	case allowModes&DisplayDefault == DisplayDefault && strings.HasSuffix(q, ss.Or(sep, `;`)):
		return q[:len(q)-1], DisplayDefault
	default:
		return q, DisplayDefault
	}
}
