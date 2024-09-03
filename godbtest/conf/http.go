package conf

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bingoohuang/ngg/godbtest/label"
	"github.com/bingoohuang/ngg/godbtest/sqlmap"
	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/ss"
)

func (c *Config) listen(wg *sync.WaitGroup) {
	defer wg.Done()
	http.HandleFunc(c.ContextPath, c.handleHTTP)
	log.Printf("start to listen and serve on %s", c.ListenAddr)
	if err := http.ListenAndServe(c.ListenAddr, nil); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen and serve on %s failed: %v", c.ListenAddr, err)
		}
	}
}

func (c *Config) handleHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	body, _ := ss.ReadAll(req.Body)
	req.Body.Close()

	q := jj.Get(body, "q").String()
	if q == "" {
		w.WriteHeader(http.StatusBadRequest)
	}

	var (
		err       error
		labelExpr *label.Visitor
	)

	labelQuery := jj.Get(body, "labelQuery").String()
	if labelQuery != "" {
		labelExpr, err = label.Parse(labelQuery)
		if err != nil {
			log.Printf("parse label %s failed: %v", labelQuery, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	offset := 0
	limit := 1000

	limitArg := jj.Get(body, "limit").String()
	if strings.Contains(limitArg, ",") {
		s0, s1 := ss.Split2(limitArg, ",")
		offset, err = strconv.Atoi(s0)
		if err != nil {
			log.Printf("bad format for limit: %s", limitArg)
		}
		limit, err = strconv.Atoi(s1)
		if err != nil {
			log.Printf("bad format for limit: %s", limitArg)
		}
	} else {
		limit, err = strconv.Atoi(limitArg)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	scanner := NewHTTPRowsScanner(offset, limit, w)

	ds := getDataSources(c.DataSources, labelExpr)
	if len(ds) == 0 {
		log.Printf("no datasource find")
		w.WriteHeader(http.StatusNotFound)
	} else if len(ds) > 1 {
		log.Printf("more than on dataource found")
		w.WriteHeader(http.StatusBadRequest)
	}
	sq := Sql{Sql: q, IgnoreError: true, SingleOne: true, RowsScanner: scanner}
	sq.Run(ds[0], &replOptions{
		limit:  limit,
		offset: offset,
	})
}

type HTTPRowsScanner struct {
	start         time.Time
	writer        io.Writer
	Header        []string
	rows          []any
	Limit, Offset int
}

func NewHTTPRowsScanner(limit, offset int, writer io.Writer) *HTTPRowsScanner {
	return &HTTPRowsScanner{Limit: limit, Offset: offset, writer: writer}
}

var _ sqlmap.RowsScanner = (*HTTPRowsScanner)(nil)

func (j *HTTPRowsScanner) StartExecute(string)                        { j.start = time.Now() }
func (j *HTTPRowsScanner) StartRows(_ string, header []string, _ int) { j.Header = header }

func (j *HTTPRowsScanner) AddRow(rowIndex int, columns []any) bool {
	if j.Offset > 0 && rowIndex < j.Offset {
		return true
	}

	if j.Limit > 0 && rowIndex+1 > j.Limit+j.Offset {
		log.Printf("Row ignored ...")
		return false
	}

	row := map[string]any{}
	for i, h := range j.Header {
		row[h] = columns[i]
	}

	j.rows = append(j.rows, row)

	return j.Limit <= 0 || rowIndex+1 <= j.Limit+j.Offset
}

type rowsResult struct {
	Cost   string   `json:"cost"`
	Header []string `json:"header"`
	Rows   []any    `json:"rows"`
}

func (j *HTTPRowsScanner) Complete() {
	cost := time.Since(j.start)
	json.NewEncoder(j.writer).Encode(
		rowsResult{
			Header: j.Header,
			Rows:   j.rows,
			Cost:   cost.String(),
		})
}
