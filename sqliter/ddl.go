package sqliter

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/bingoohuang/ngg/sqliter/influx"
	"github.com/bingoohuang/ngg/ss"
	"github.com/samber/lo"
)

// CreateCreateTable 创建 建表 SQL, 索引 SQL
func CreateCreateTable(metric influx.Metric, seq *BoltSeq, asTags Matcher) (tableColumns, scripts []string) {
	t := metric.Name()

	tableColumns = []string{"timestamp"}
	pks := []string{"timestamp"}
	cols := []string{`"timestamp" text`}
	// 存放创建索引的语句
	var idx []string

	for c := range metric.Tags() {
		idx = append(idx, createIndex(t, c))
		tableColumns = append(tableColumns, c)
		pks = append(pks, c)
		col := strconv.Quote(c) + ss.If(seq != nil, " NUMERIC", " TEXT")
		cols = append(cols, col)
	}

	for c, v := range metric.Fields() {
		tableColumns = append(tableColumns, c)
		col := strconv.Quote(c) + " " + deriveDataType(c, v, asTags, seq)
		cols = append(cols, col)

		if asTags.Match(c) {
			idx = append(idx, createIndex(t, c))
			pks = append(pks, c)
		}
	}

	scripts = []string{
		fmt.Sprintf("drop table if exists %q", t),
		fmt.Sprintf("create table %q(%s)", t, strings.Join(cols, ",")),
	}

	if len(pks) == 1 {
		scripts = append(scripts, createUniqueIndex(t, "timestamp"))
	} else {
		scripts = append(scripts,
			createIndex(t, "timestamp"),
			createUniqueIndex(t, pks...))
	}

	return tableColumns, append(scripts, idx...)
}

// CreateAlterTable 根据指标，生成添加字段、创建索引等 SQL 语句
func CreateAlterTable(ti *tableMeta, metric influx.Metric, seq *BoltSeq, asTags Matcher) []string {
	t := metric.Name()

	var alters []string
	hasNewIndex := false

	// 对 Tags 增加 表列 及 索引
	for c := range metric.Tags() {
		if _, ok := ti.Headers[c]; ok {
			continue
		}

		typ := ss.If(seq != nil, "NUMERIC", "TEXT")
		alter := fmt.Sprintf("alter table %q add %q %s", t, c, typ)
		alters = append(alters, alter, createIndex(t, c))
		hasNewIndex = true
		ti.Tags[c] = true
		ti.Headers[c] = true
	}

	// 对 Fields  增加表列，及 可选索引 (由 asTags 确定)
	for c, v := range metric.Fields() {
		if _, ok := ti.Headers[c]; ok {
			continue
		}

		alters = append(alters, fmt.Sprintf("alter table %q add %q %s",
			t, c, deriveDataType(c, v, asTags, seq)))

		ti.Headers[c] = true
		if asTags.Match(c) { // 字段作为 TAG, 需要创建对应的索引
			alters = append(alters, createIndex(t, c))
			hasNewIndex = true
			ti.Tags[c] = true
		}
	}

	if hasNewIndex {
		alters = append(alters,
			dropUniqueIndex(t),
			createUniqueIndex(t, lo.Keys(ti.Tags)...))
	}

	return alters
}

func createIndex(table string, column ...string) string {
	parts := []string{"idx", table}
	parts = append(parts, column...)

	quotedCols := lo.Map(column, func(c string, _ int) string { return strconv.Quote(c) })

	idx := strconv.Quote(strings.Join(parts, "_"))
	joinedCols := strings.Join(quotedCols, ",")
	return fmt.Sprintf("create index %s on %q(%s)", idx, table, joinedCols)
}

func dropUniqueIndex(table string) string {
	return fmt.Sprintf("drop index %q", "udx_"+table)
}

func createUniqueIndex(table string, column ...string) string {
	sortedColumns := uniqueIndexSort(column)
	quotedCols := lo.Map(sortedColumns, func(c string, _ int) string { return strconv.Quote(c) })
	joinedCols := strings.Join(quotedCols, ",")
	return fmt.Sprintf("create unique index %q on %q(%s)", "udx_"+table, table, joinedCols)
}

// uniqueIndexSort 排序：将 "timestamp" 放在第1位，其余按字典顺序排序
func uniqueIndexSort(data []string) (result []string) {
	var others []string

	for _, s := range data {
		if s == "timestamp" {
			result = append(result, s)
		} else {
			others = append(others, s)
		}
	}

	// 按字母顺序排序其余元素
	sort.Strings(others)

	// 将 timestamp 插入到第1位
	return append(result, others...)
}
