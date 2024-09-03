package conf

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/bingoohuang/ngg/godbtest/sqlmap"
	"github.com/bingoohuang/ngg/ss"
	"github.com/samber/lo"
	"github.com/spf13/pflag"
)

func init() {
	registerOptions(`%ants`,
		`
蚂蚁搬家式的更新大表，批次大小100，前10条/每10s/每更新1000条，打印日志

%ants -N 100 --log 10p,10s,1000r --query='select ID from zz where MODIFIED < '2023-05-17' [and ID >= :ID] order by ID limit :N' --update='update zz set MODIFIED = now() where ID in (:ID{N})'

`,
		func(name string, options *replOptions) {
			log.Printf("%s -N %d --log %s %s", name, options.antN, options.antLog, lo.Ternary(options.dryRun, " --dry ", ""))
			log.Printf("%s --query %s", name, ss.QuoteSingle(options.antQuery))
			log.Printf("%s --update %s", name, ss.QuoteSingle(options.antUpdate))
			log.Printf("antState: %s", ss.Json(options.antState))
		}, parseAntsArgs)
}

func parseAntsArgs(name string, opt *replOptions, args []string, pureArg string) error {
	opt.antClear = false

	f := pflag.NewFlagSet(name, pflag.ContinueOnError)
	f.IntVarP(&opt.antMax, "max", "M", 0, "最多更新多少条")
	f.IntVarP(&opt.antN, "num", "N", 1000, "批量大小")
	f.StringVarP(&opt.antLog, "log", "l", "10p,10s,1000r", "前10条/每10秒/每更新1000条，打印日志")
	f.StringVarP(&opt.antQuery, "query", "q", "", "查询语句")
	f.StringVarP(&opt.antUpdate, "update", "u", "", "更新语句")
	f.StringVarP(&opt.antStateFlag, "state", "", "", `状态 JSON, 例如 '{"ID": 100}''`)
	f.BoolVarP(&opt.dryRun, "dry", "", false, "空跑，不更新")
	f.BoolVarP(&opt.antClear, "clear", "", false, "清空状态")
	f.BoolVarP(&opt.antConfirm, "confirm", "c", false, "是否确认继续运行")
	if err := f.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	if opt.antStateFlag != "" {
		if err := json.Unmarshal([]byte(opt.antStateFlag), &opt.antState); err != nil {
			return fmt.Errorf("unmarshal state: %w", err)
		}
	}

	return opt.RunAnts()
}

func (o *replOptions) RunAnts() error {
	if len(o.dss) == 0 {
		return fmt.Errorf("no data sources specified, please type %%help to get help")
	}

	if o.antQuery == "" {
		return fmt.Errorf("%%ant --query should be set")
	}

	ds := o.dss[0]

	if o.antClear || len(o.antState) == 0 {
		o.antState = map[string]any{}
	}

	f := &FindOption{Open: '[', Close: ']', Quote: '\''}
	queryTags := f.FindTags(o.antQuery)
	var query string
	if len(o.antState) == 0 {
		query = FirstTimeQuery(o.antQuery, queryTags)
	} else {
		query = FullQuery(o.antQuery, queryTags)
	}

	named := f.FindNamed(query)
	o.antState["N"] = o.antN

	placeholderFn := GetBindPlaceholder(ds.DriverName)
	bindQuery, bindArgs, err := parseBindArgs(query, named, o.antN, placeholderFn)
	if err != nil {
		return err
	}

	totalExpected := 0
	realAffected := int64(0)

	var (
		bindUpdate string
		updateArgs []bindArg
	)

	if o.antUpdate != "" {
		if bindUpdate, updateArgs, err = parseBindArgs(o.antUpdate, f.FindNamed(o.antUpdate), o.antN, placeholderFn); err != nil {
			return err
		}
		defer func() {
			log.Printf("last state: %s", ss.Json(o.antState))
		}()
	} else {
		o.antLog = "all"
	}

	antLogOpt, err := ParseLogOptions(o.antLog)
	if err != nil {
		return fmt.Errorf("parse log options %s: %w", o.antLog, err)
	}
	start := time.Now()
	logState := &LogState{Opt: antLogOpt}
	r := antsRuns{
		query:         query,
		bindArgs:      bindArgs,
		bindQuery:     bindQuery,
		updateArgs:    updateArgs,
		bindUpdate:    bindUpdate,
		placeholderFn: placeholderFn,
		queryTags:     queryTags,
		logState:      logState,
		namedPos:      named,
		findOption:    f,
		logger:        lo.Ternary(o.antConfirm, log.Printf, logState.Log),
	}

	for batch := 0; ; batch++ {
		if err := r.doBatch(o, ds, batch); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
	}

	cost := time.Since(start)
	log.Printf("Complete total affected: %d/%d, cost: %s", realAffected, totalExpected, cost)

	return nil
}

type antsRuns struct {
	placeholderFn func(seq int) string
	logger        func(format string, v ...any)
	logState      *LogState
	findOption    *FindOption
	bindQuery     string
	query         string
	bindUpdate    string
	bindArgs      []bindArg
	updateArgs    []bindArg
	queryTags     []*Pos
	namedPos      []*Pos
	realAffected  int64
	totalExpected int
}

func (r *antsRuns) doBatch(o *replOptions, ds *DataSource, batch int) error {
	queryStart := time.Now()
	args := createBindArgs(r.bindArgs, o.antState)
	ctx := context.Background()
	ctx, cancelFn := WithSignal(ctx, o.timeout)
	defer cancelFn()

	queryRows, err := sqlmap.Select(ctx, ds.db, r.bindQuery, args...)
	if err != nil {
		return fmt.Errorf("query %q: %w", r.bindQuery, err)
	}
	queryCost := time.Since(queryStart)

	r.logger("Query#%03d %q :: %v result %d rows, cost: %s", batch+1, r.bindQuery, args, len(queryRows), queryCost)

	if len(queryRows) == 0 {
		return io.EOF
	}

	queryRowsStepper := &QueryRowsStepper{Rows: queryRows}
	if queryRowsStepper.HasNext() {
		if o.antUpdate == "" {
			queryRowsStepper.Step(o.antState)
		} else if err := r.doUpdate(o, ds, queryRowsStepper); err != nil {
			return err
		}
	}

	if o.antMax > 0 && r.totalExpected >= o.antMax {
		return io.EOF
	}

	if o.antConfirm {
		continueUpdate := true
		prompt := &survey.Confirm{Message: "Continue?", Default: true}
		if err := survey.AskOne(prompt, &continueUpdate); err != nil {
			return err
		}
		if !continueUpdate {
			return io.EOF
		}
	}

	if batch == 0 {
		r.query = FullQuery(o.antQuery, r.queryTags)
		r.namedPos = r.findOption.FindNamed(r.query)
		r.bindQuery, r.bindArgs, err = parseBindArgs(r.query, r.namedPos, o.antN, r.placeholderFn)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *antsRuns) doUpdate(o *replOptions, ds *DataSource, queryRowsStepper *QueryRowsStepper) error {
	updateStart := time.Now()
	scanner := NewUpdateRowsScanner()
	updateBindArgs, expectedRows := createUpdateBindArgs(r.updateArgs, queryRowsStepper, o.antState)
	ctx := context.Background()
	ctx, cancelFn := WithSignal(ctx, o.timeout)
	defer cancelFn()

	if err := Exec(ctx, ds.db, r.bindUpdate, updateBindArgs, scanner); err != nil {
		return fmt.Errorf("update %q with args %v: %w", r.bindUpdate, updateBindArgs, err)
	}
	updateCost := time.Since(updateStart)
	r.realAffected += scanner.affected
	r.totalExpected += expectedRows
	r.logState.AddRows(expectedRows)
	r.logger("Update %q :: %v affected: %d/%d, total: %d/%d, cost: %s",
		r.bindUpdate, updateBindArgs, scanner.affected, expectedRows, r.realAffected, r.totalExpected, updateCost)
	return nil
}

type QueryRowsStepper struct {
	Rows  []map[string]any
	Index int
}

func (s *QueryRowsStepper) Step(state map[string]any) bool {
	if s.Index >= len(s.Rows) {
		return false
	}

	for k, v := range s.Rows[s.Index] {
		state[k] = v
	}
	s.Index++
	return true
}

func (s *QueryRowsStepper) HasNext() bool {
	return s.Index < len(s.Rows)
}

type updateRowsScanner struct {
	insertId int64
	affected int64
}

func (n *updateRowsScanner) StartExecute(string)             {}
func (n *updateRowsScanner) StartRows(string, []string, int) {}
func (n *updateRowsScanner) Complete()                       {}
func (n *updateRowsScanner) AddRow(_ int, row []any) bool {
	n.insertId = row[0].(int64)
	n.affected = row[1].(int64)
	return true
}

func NewUpdateRowsScanner() *updateRowsScanner { return &updateRowsScanner{} }

func createUpdateBindArgs(args []bindArg, stepper *QueryRowsStepper, state map[string]any) (bindArgs []any, expectedRows int) {
	bindArgs = make([]any, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg.Repeat == 0 {
			bindArgs[i] = state[arg.Name]
			expectedRows++
			continue
		}

		if stepper.Step(state) {
			bindArgs[i] = state[arg.Name]
			expectedRows++
		}

		for i++; i < len(args); i++ {
			if args[i].Repeat == 0 || args[i].Repeat == 1 {
				i--
				break
			}

			if stepper.Step(state) {
				bindArgs[i] = state[arg.Name]
				expectedRows++
			}
		}
	}
	return bindArgs, expectedRows
}

func createBindArgs(args []bindArg, state map[string]any) []any {
	bindArgs := make([]any, len(args))
	for i, arg := range args {
		bindArgs[i] = state[arg.Name]
	}

	return bindArgs
}

type bindArg struct {
	Name   string
	Arg    string
	Repeat int
}

func parseBindArgs(query string, named []*Pos, antN int, placeholderFn func(seq int) string) (bindQuery string, bindArgs []bindArg, err error) {
	p := 0
	seq := 0
	for _, name := range named {
		arg := bindArg{}
		if name.ArgOpen == 0 {
			arg.Name = query[name.From+1 : name.To]
		} else {
			arg.Name = query[name.From+1 : name.ArgOpen]
		}

		bindQuery += query[p:name.From]
		seq++
		if name.ArgOpen == 0 {
			bindQuery += placeholderFn(seq)
			arg.Arg = query[name.From:name.To]
			bindArgs = append(bindArgs, arg)
		} else {
			nameArg := strings.TrimSpace(query[name.ArgOpen+1 : name.To-1])
			nameArg = strings.ToUpper(nameArg)

			if nameArg != "N" {
				antN, err = strconv.Atoi(nameArg)
				if err != nil || antN < 0 {
					return "", nil, fmt.Errorf("bad arg %s", query[name.From:name.To])
				}
			}

			placeHolders := make([]string, antN)
			for j := 0; j < antN; j++ {
				placeHolders[j] = placeholderFn(seq)
				seq++

				arg.Repeat = j + 1
				bindArgs = append(bindArgs, arg)
			}
			bindQuery += strings.Join(placeHolders, ",")
		}

		p = name.To
	}

	bindQuery += query[p:]
	return
}

func FullQuery(query string, tags []*Pos) string {
	var q string
	p := 0
	for _, t := range tags {
		q += query[p:t.From] + query[t.From+1:t.To-1]
		p = t.To
	}

	q += query[p:]
	return q
}

func FirstTimeQuery(query string, tags []*Pos) string {
	var q string
	p := 0
	for _, t := range tags {
		q += query[p:t.From]
		p = t.To
	}

	q += query[p:]
	return q
}

type LogState struct {
	Last     time.Time
	Opt      *LogOption
	Rows     int
	LogTimes int
}

func (s *LogState) AddRows(rows int) {
	s.Rows += rows
}

func (s *LogState) Log(fmt string, args ...any) {
	s.LogTimes++
	if s.Opt.All ||
		s.LogTimes <= s.Opt.Preview ||
		s.Opt.Rows > 0 && s.Rows >= s.Opt.Rows ||
		s.Opt.Interval > 0 && time.Since(s.Last) >= s.Opt.Interval {
		log.Printf(fmt, args...)

		s.Rows = 0
		s.Last = time.Now()
	}
}

type LogOption struct {
	Interval time.Duration
	Preview  int
	Rows     int
	All      bool
}

func ParseLogOptions(option string) (logOpt *LogOption, err error) {
	parts := strings.Split(option, ",")
	logOpt = &LogOption{}
	for _, part := range parts {
		if strings.ToLower(part) == "all" {
			logOpt.All = true
			break
		}
		switch lastLetter := part[len(part)-1]; {
		case lastLetter >= '0' && lastLetter <= '9':
			logOpt.Rows, err = strconv.Atoi(part)
			if err != nil {
				return nil, err
			}
		case lastLetter == 'p':
			logOpt.Preview, err = strconv.Atoi(part[:len(part)-1])
			if err != nil {
				return nil, err
			}
		case lastLetter == 'r':
			logOpt.Rows, err = strconv.Atoi(part[:len(part)-1])
			if err != nil {
				return nil, err
			}
		default:
			logOpt.Interval, err = time.ParseDuration(part)
			if err != nil {
				return nil, err
			}
		}
	}

	return logOpt, nil
}
