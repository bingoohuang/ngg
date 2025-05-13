package sqlmap

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/alecthomas/chroma/quick"
	"github.com/bingoohuang/ngg/godbtest/drivers"
	"github.com/bingoohuang/ngg/ss"
	"github.com/xo/dburl"
)

type Config struct {
	MaxOpenConns int
	Verbose      bool
}

type ConfigFn func(*Config)

func WithMaxOpenConns(maxOpenConns int) ConfigFn {
	return func(c *Config) {
		c.MaxOpenConns = maxOpenConns
	}
}

func WithVerbose(verbose bool) ConfigFn {
	return func(c *Config) {
		c.Verbose = verbose
	}
}

var optionMap = ss.SplitToMap(os.Getenv("GODBTEST_OPTIONS"), ",", ":")

type OptionKey string

const (
	OptionNoPing OptionKey = "noping"
)

func (k OptionKey) Value() string { return optionMap[string(k)] }
func (k OptionKey) Exists() bool  { _, ok := optionMap[string(k)]; return ok }

// Connect 创建数据库连接
func Connect(driverName, dataSourceName string, options ...ConfigFn) (*sql.DB, error) {
	var config Config
	for _, option := range options {
		option(&config)
	}

	switch driverName {
	case "sqlite", "sqlite3":
		dataSourceName = ss.ExpandHome(dataSourceName)
	}

	u, err := url.Parse(dataSourceName)
	if err != nil {
		return nil, err
	}
	dbFixer := drivers.FixURL(u)
	db, err := OpenDB(u.String(), config.Verbose)
	if err != nil {
		return nil, err
	}

	if dbFixer != nil {
		if err := dbFixer(db); err != nil {
			return nil, fmt.Errorf("dbFixer: %w", err)
		}
	}

	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
	}

	if !OptionNoPing.Exists() {
		if err := db.Ping(); err != nil {
			db.Close()
			return nil, fmt.Errorf("ping: %w", err)
		}
	}

	if config.Verbose {
		log.Printf("connected to %s succeed", dataSourceName)
	}
	return db, nil
}

func OpenDB(urlstr string, verbose bool) (*sql.DB, error) {
	u, err := dburl.Parse(urlstr)
	if err != nil {
		return nil, err
	}
	driver := u.Driver
	if u.GoDriver != "" {
		driver = u.GoDriver
	}
	if driver == "sqlite3" {
		driver = "sqlite"
	}

	if verbose {
		log.Printf("connecting to driver: %s, DSN: %s ...", driver, u.DSN)
	}

	return sql.Open(driver, u.DSN)
}

// Color highlight the SQL in color.
func Color(q string) string {
	buf := bytes.NewBuffer([]byte{})
	err := quick.Highlight(buf, q, "sql", "terminal16m", "monokai")
	if err != nil {
		log.Fatal(err)
	}
	return buf.String()
}
