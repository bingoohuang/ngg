package conf

import (
	"database/sql"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/ss"

	"github.com/spf13/pflag"
	"github.com/xo/dburl"
)

type replOptions struct {
	lookups map[string]map[string]string

	antState            map[string]any
	sep                 string
	format              string
	thinkTime           string
	antLog              string
	antQuery, antUpdate string
	antStateFlag        string
	rawFileDir          string
	rawFileExt          string
	dss                 []*DataSource
	perfDuration        time.Duration
	verbose             int
	times               int

	antN, antMax int

	num, threads, batch int

	limit, offset int
	maxLen        int
	lookupsOn     bool
	showCost      bool

	antClear, antConfirm bool

	named bool

	dryRun          bool
	asQuery         bool
	asExec          bool
	showDrivers     bool
	noParsePrepared bool
	clearSetting    bool

	printSQL     bool
	ShowRowIndex bool
	NoEvalSQL    bool

	saveResult   bool
	tempResultDB bool

	timeout time.Duration // SQL 执行超时时间

	qps float64

	maxOpenConns int
}

func parseSetArgs(name string, options *replOptions, args []string, pureArg string) error {
	f := pflag.NewFlagSet(name, pflag.ContinueOnError)
	f.IntVarP(&options.times, "times", "", 1, "执行次数")

	f.IntVarP(&options.offset, "offset", "", 0, "偏移量，与 limit 结合使用")
	f.IntVarP(&options.maxLen, "maxLen", "", 32, "最大显示长度，超长…")
	f.IntVarP(&options.limit, "limit", "", 1000, "最大查询行数")
	f.StringVarP(&options.sep, "sep", "", ";", "语句分隔符")
	f.StringVarP(&options.format, "format", "", "", "结果显示格式, none/table/json/json:free/markdown/csv/insert")
	f.BoolVarP(&options.named, "named", "", false, "执行配置文件中预备的命名语句")
	f.BoolVarP(&options.showCost, "cost", "", false, "显示执行时间")
	f.BoolVarP(&options.saveResult, "saveResult", "", false, "自动保存执行结果")
	f.BoolVarP(&options.tempResultDB, "db", "", false, "自动保存执行结果到sqlite表中")
	f.DurationVarP(&options.timeout, "timeout", "", 0, "SQL执行超时时间, e.g. 30s")
	f.StringVarP(&options.rawFileDir, "writeLob", "", "", "保存lob到目录中，e.g. /tmp")
	f.StringVarP(&options.rawFileExt, "lobExt", "", "", "lob 后缀名，e.g. .jpg")
	f.BoolVarP(&options.ShowRowIndex, "showNumber", "", false, "结果添加行索引号")
	f.BoolVarP(&options.NoEvalSQL, "noEvalSQL", "", false, "不求值 SQL 表达式")
	f.BoolVarP(&options.clearSetting, "clear", "", false, "清空设置")
	f.Float64Var(&options.qps, "qps", 0, "QPS rate limit")

	f.IntVarP(&options.maxOpenConns, "maxOpenConns", "", 0, "db.SetMaxOpenConns")
	f.IntVarP(&options.num, "num", "n", 0, "total insert records/update operations")
	f.IntVarP(&options.threads, "threads", "t", runtime.NumCPU(), "number of goroutines to query")
	f.DurationVarP(&options.perfDuration, "duration", "d", 0, "duration to run the performance test")
	f.IntVarP(&options.batch, "batch", "b", 100, "batch size to generate")
	f.CountVarP(&options.verbose, "verbose", "v", "verbose \n"+
		"-v: print bind values line by line \n"+
		"-vv: print bind values in tab by tab")
	f.BoolVarP(&options.dryRun, "dry", "", false, "dry run without execution")
	f.BoolVarP(&options.asQuery, "asQuery", "Q", false, "execute as query")
	f.BoolVarP(&options.asExec, "asExec", "E", false, "execute as exec")
	f.BoolVarP(&options.showDrivers, "drivers", "", false, "显示所有支持的数据库驱动名")
	f.BoolVarP(&options.noParsePrepared, "noPrepared", "", false, "do not parse prepared SQL")
	f.StringVarP(&options.thinkTime, "think", "k", "", "think time among requests, eg. 1s, 10ms, 10-20ms. (unit ns, us/µs, ms, s, m, h)")
	err := f.Parse(args)
	if err == nil {
		if options.showDrivers {
			fmt.Printf("Dirvers: %v\n", sql.Drivers())
		}
	}
	return err
}

func init() {
	registerOptions(`%help`, `%help {cmd};`,
		func(name string, options *replOptions) {
			fmt.Println("@test.sql # 直接执行 SQL 文件")
			fmt.Println("SQL;  # 普通输出模式")
			fmt.Println(`SQL\G # 纵向打印数据行（每行一个列值）`)
			fmt.Println(`SQL\I # Insert SQL 打印数据行`)
			fmt.Println(`SQL\J # JSON 打印数据行`)
			fmt.Println(`SQL\P # Performance 性能压测模式`)
			fmt.Printf("%%begin/%%commit/%%rollback\n")
			fmt.Println(getOptionUsages("", "help"))
		}, func(name string, options *replOptions, args []string, pureArg string) error {
			fmt.Println(getOptionUsages(args[0], "help"))
			return nil
		})

	registerOptions(`%set`, `%set -times 1 -verbose 1
--times {N}      执行 N 次
--verbose {N}    反显级别
--offset {N}     偏移量，与 limit 结合使用
--limit {count}  最大查询行数
--sep   {sep}    语句分隔符
--format {fmt}   结果显示格式, none/table/json/json:free/markdown/csv/insert
--named          执行配置文件中预备的命名语句
--cost           显示执行时间
--saveResult     自动保存执行结果
--timeout {时间段} SQL执行超时时间
--writeLob {dir}  保存lob到目录中，e.g. /tmp
--lobExt  {.ext}  lob 后缀名，e.g. .jpg
--showNumber      结果添加行索引号
--noEvalSQL       不求值 SQL 表达式
--noPrepared      不解析 SQL 表达式中的常量
--dry             干跑，不执行
--clear           清空所有设置到默认值
--drivers         显示所有支持的数据库驱动名

%set -n {TotalOperationNum, 默认 1000} -t {threadsNum 默认核数} -b {batchNum 默认100} -v (不输出进度条，打印绑定变量) -think {3s};
示例:
> %connect 'oracle://system:Password123@172.16.177.145:1521/xkdg?lob fetch=post'
> CREATE TABLE t2 (id char(27) primary key, name varchar(255), addr varchar(255), email varchar(255), phone varchar(255), age varchar(3), idc varchar(256)) engine=innodb default charset=utf8mb4;
> %set -n 10000;
> insert into t2(id,name,addr,email,phone,age,idc) values('@ksuid','@姓名','@地址','@邮箱','@手机','@random_int(15-95)','@身份证')\P

可以使用的插值表达式：
1. @uuid @ksuid @name @address @state
2. @汉字 @姓名 @性别 @地址 @手机 @身份证 @发证机关 @邮箱 @银行卡
3. @random(5-10) @random(red,green,blue)  @random(1,2,3)
4. @regex([abc]{10}) @regex([a-z]{5}@xyz[.]cn)
5. @random_int" @random_int(100-999)" @random_bool")
6. @random_time" @random_time(yyyy-MM-dd) @random_time(yyyy-MM-ddTHH:mm:ss)
7. @random_time(yyyy-MM-dd,1990-01-01,2021-06-06) @random_time(sep=# yyyy-MM-dd#1990-01-01#2021-06-06)
8. @seq create sequence starting from 1, or @seq(100)  create sequence starting from 100.
9. @file(/home/1.jpg,:bytes)
10. @random_image(format=jpg size=640x320)  @base64(size=1000 std raw file=dir/f.png)
11. @emoji @emoji(3) @emoji(3-5)
`,
		func(name string, options *replOptions) {
			log.Printf("%s -times %d -verbose %d -offset %d -limit %d -sep %s -format %s "+
				"-named %t -cost %t -saveResult %t -timeout %s -writeLob %s -lobExt %s"+
				"-showNumber %t -evalSQL %t", name,
				options.times, options.verbose, options.offset, options.limit, options.sep, options.format,
				options.named, options.showCost, options.saveResult, options.timeout, options.rawFileDir, options.rawFileExt,
				options.ShowRowIndex,
				!options.NoEvalSQL,
			)
		}, parseSetArgs)

	registerOptions(`%connect`, `%connect 'dm://SYSDBA:123456@127.0.0.1:5236?schema=demo';
%connect 'pgx://user:pass@127.0.0.1:54321/demo?sslmode=disable';

dburl examples:
postgres://user:pass@localhost/dbname
pg://user:pass@localhost/dbname?sslmode=disable
mysql://user:pass@localhost/dbname?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true
mysql:/var/run/mysqld/mysqld.sock
sqlserver://user:pass@remote-host.com/dbname
mssql://user:pass@remote-host.com/instance/dbname
ms://user:pass@remote-host.com:port/instance/dbname?keepAlive=10
oracle://user:pass@somehost.com/sid?lob fetch=post
sap://user:pass@localhost/dbname
sqlite:/path/to/file.db
sqlite:/:memory:
file:myfile.sqlite3?loc=auto
odbc+postgres://user:pass@localhost:port/dbname?option1=

or use ENV vars like:

export DSN=sqlite/:memory:
export DSN=mysql://root:root@127.0.0.1:3306/bingoo
export DSN=oracle://user:pass@127.0.0.1:1521/ORCLPDB1?lob fetch=post
`,
		func(name string, options *replOptions) {
			log.Printf(`%%connect driverName datasourceName;`)
		}, func(name string, options *replOptions, args []string, pureArg string) error {
			if pureArg == "" {
				pureArg = args[0]
			}
			if pureArg != "" {
				u, err := dburl.Parse(pureArg)
				if err != nil {
					return err
				}
				options.dss = connectDB(u.Driver, pureArg, options.dss, options.verbose, options.maxOpenConns)
			}

			return nil
		})

	registerOptions(`%close`, `%close`, func(name string, options *replOptions) {
		ss.Close(options.dss...)
		options.dss = nil
	}, func(name string, options *replOptions, args []string, pureArg string) error {
		ss.Close(options.dss...)
		options.dss = nil
		return nil
	})
}

type optionFns struct {
	getter func(name string, options *replOptions)
	setter func(name string, options *replOptions, args []string, pureArg string) error
	name   string
	usage  string
}

var optionFnsRegister = make(map[string]*optionFns)

func registerOptions(name, usage string,
	getter func(name string, options *replOptions),
	setter func(name string, options *replOptions, args []string, pureArg string) error,
) {
	optionFnsRegister[name] = &optionFns{
		name:   name,
		usage:  usage,
		getter: getter,
		setter: setter,
	}
}

func getOptionUsages(with, without string) string {
	usage := ""
	if with != "" && !strings.HasPrefix(with, "%") {
		with = "%" + with
	}
	for _, f := range optionFnsRegister {
		if without != "" && strings.EqualFold(f.name, without) {
			continue
		}
		if with != "" && !strings.EqualFold(f.name, with) {
			continue
		}

		if usage != "" && !strings.HasSuffix(usage, "\n") {
			usage += "\n"
		}
		usage += f.usage
	}
	return usage
}

func findOptionFunc(name string) *optionFns {
	for n, fns := range optionFnsRegister {
		if strings.EqualFold(n, name) {
			return fns
		}
	}

	return nil
}
