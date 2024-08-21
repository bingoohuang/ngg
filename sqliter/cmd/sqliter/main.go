package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bingoohuang/ngg/sqliter"
	"github.com/bingoohuang/ngg/sqliter/influx"
	"github.com/bingoohuang/ngg/ver"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nxadm/tail"
)

func main() {
	query := flag.String("query", "SELECT sqlite_version()", "sql query")
	tailFile := flag.String("tail", "", "json log file to tail")
	prefix := flag.String("prefix", "sqliter.t", "sqlite db file prefix (with path), e.g. /etc/sqliter/sqliter.t")
	debug := flag.Bool("debug", false, "debug mode, more logging")
	version := flag.Bool("version", false, "show version, then exit")
	follow := flag.Bool("follow", false, "follow tail")
	reopen := flag.Bool("reopen", false, "reopen tail")
	flag.Parse()

	if *version {
		fmt.Printf("%s\n", ver.Version())
		return
	}

	if *tailFile == "" {
		log.Fatalf("tail argument required")
	}

	// Create a tail
	t, err := tail.TailFile(*tailFile, tail.Config{Follow: *follow, ReOpen: *reopen})
	if err != nil {
		panic(err)
	}

	plus, err := sqliter.New(
		sqliter.WithDriverName("sqlite3"),
		sqliter.WithPrefix(*prefix),
		sqliter.WithDebug(*debug),
	)
	if err != nil {
		panic(err)
	}
	defer plus.Close()

	lineDone := make(chan struct{})

	var table string
	var tt time.Time

	go func() {
		var (
			p   *influx.Point
			err error
		)
		lines := 0
		start := time.Now()
		// Print the text of each received line
		for line := range t.Lines {
			if strings.HasPrefix(line.Text, "{") {
				p = &influx.Point{}
				err = json.Unmarshal([]byte(line.Text), &p)
				p.MetricTime = time.Unix(p.Timestamp, 0)
			} else {
				p, err = influx.ParseLineProtocol(line.Text)
			}

			if err != nil {
				log.Fatalf("parse line: %d error: %v", line.Num, err)
			}
			table = p.Name()
			tt = p.Time()

			if err := plus.WriteMetric(p); err != nil {
				log.Fatalf("write metric error: %v", err)
			}
			lines++

			if time.Since(start) > 10*time.Second {
				log.Printf("total lines: %d processed", lines)
				start = time.Now()
			}
		}

		log.Printf("total lines: %d processed", lines)
		lineDone <- struct{}{}
	}()

	// 创建一个信号通道
	sigs := make(chan os.Signal, 1)
	// 监听 SIGINT 信号，Ctrl+C 会触发这个信号
	signal.Notify(sigs, syscall.SIGINT)
	// 阻塞等待信号
	select {
	case sig := <-sigs:
		log.Printf("received singal %s", sig)
	case <-lineDone:
	}

	if *query != "" {
		result, err := plus.Read(table, *query, tt, nil)
		if err != nil {
			log.Printf("error: %v", err)
		} else {
			log.Printf("result: %+v", result)
		}
	}

	if err := t.Stop(); err != nil {
		log.Printf("stop tail error: %v", err)
	}
}
