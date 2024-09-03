package conf

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/bingoohuang/ngg/godbtest/files"
	"github.com/bingoohuang/ngg/godbtest/sqlmap"
	"github.com/bingoohuang/ngg/ss"
)

func init() {
	registerOptions(`%loglinetest`, `%loglinetest {line}`,
		func(name string, options *replOptions) {},
		func(name string, options *replOptions, args []string, pureArg string) error {
			if len(args) > 0 {
				m := files.ParseLogLine(args[0])
				log.Printf("%s", ss.Json(m))
			}

			return nil
		})

	registerOptions(`%loglinedb`, `%loglinedb file.txt`,
		func(name string, options *replOptions) {},
		func(name string, options *replOptions, args []string, pureArg string) error {
			if len(args) == 0 {
				return fmt.Errorf("usage: %%loglinedb file.txt")
			}

			return parseLogFileToDB(args[0])
		})
}

func parseLogFileToDB(file string) error {
	base := filepath.Base(file)
	tt := time.Now().Format(`_20060102150405`)
	dbFile := strings.TrimSuffix(base, filepath.Ext(base)) + tt + ".db"
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("open db file: %w", err)
	}
	defer ss.Close(db)

	lineCh := make(chan string)
	errCh := make(chan error)
	go files.GetLine(file, lineCh, errCh)

	prompt := ""
	tableName := "log" + tt
	f := func(b *sqlmap.BatchUpdate, start, batchStart time.Time, totalNum, batchNum int, complete bool) {
		if complete {
			cost := time.Now().Sub(start)
			log.Printf("file %q table %q (%d rows) in %s (avg %s/row) complete!",
				dbFile, tableName, totalNum, cost, cost/time.Duration(totalNum))
		} else if cost := time.Now().Sub(batchStart); cost > 3*time.Second {
			if clearLine := ss.Repeat("\b", "", len(prompt)); clearLine != "" {
				fmt.Print(clearLine)
			}
			prompt = fmt.Sprintf("%d rows in %s", batchNum, cost)
		}
	}

	var keys []string
	var batchUpdate *sqlmap.BatchUpdate

	for line := range lineCh {
		lineMap := files.ParseLogLine(line)
		if batchUpdate == nil {
			var fieldsKeys []string
			var fields []string
			for _, k := range lineMap.Keys() {
				keys = append(keys, k.(string))

				ks := k.(string)
				ks = strings.ReplaceAll(ks, "-", "_")
				ks = strings.ReplaceAll(ks, ".", "_")
				fieldsKeys = append(fieldsKeys, ks)
				fields = append(fields, ks+" text")
			}

			ct := `CREATE TABLE ` + tableName + `(` + strings.Join(fields, ", ") + `)`
			if _, err := db.Exec(ct); err != nil {
				return fmt.Errorf("create table %s: %w", tableName, err)
			}

			q := `INSERT INTO ` + tableName + `(` + strings.Join(fieldsKeys, ", ") + `) VALUES(` + ss.Repeat("?", ",", len(keys)) + `)`
			batchUpdate = sqlmap.NewBatchUpdate(db, q, sqlmap.WithBatchNotifier(f))
		}

		var values []any
		for _, k := range keys {
			value, _ := lineMap.Get(k)
			values = append(values, value)
		}

		batchUpdate.AddVars(values)
	}

	return batchUpdate.Close()
}
