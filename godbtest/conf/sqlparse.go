package conf

import (
	"fmt"
	"log"
	"strings"

	"github.com/bingoohuang/ngg/sqlparser"
	"github.com/bingoohuang/ngg/sqlparser/dialect"
	"github.com/bingoohuang/ngg/sqlparser/dialect/mysql"
	"github.com/bingoohuang/ngg/sqlparser/dialect/postgresql"
	"github.com/bingoohuang/ngg/ss"
)

type BindNameAware interface {
	CurrentSeq() int
	NextBindName() string
	BindName(seq int) string
	GetDialect() dialect.Dialect
}

type Seq struct {
	seq int
}

func (s *Seq) CurrentSeq() int {
	return s.seq
}

type BindNameOracle struct {
	Seq
}

func (s *BindNameOracle) GetDialect() dialect.Dialect {
	return postgresql.NewPostgreSQLDialect()
}

func (s *BindNameOracle) NextBindName() string {
	s.seq++
	return fmt.Sprintf(":%d", s.seq)
}

func (s *BindNameOracle) BindName(seq int) string {
	return fmt.Sprintf(":%d", seq)
}

type BindNamePg struct {
	Seq
}

func (s *BindNamePg) GetDialect() dialect.Dialect {
	return postgresql.NewPostgreSQLDialect()
}

func (s *BindNamePg) NextBindName() string {
	s.seq++
	return fmt.Sprintf("$%d", s.seq)
}

func (s *BindNamePg) BindName(seq int) string {
	return fmt.Sprintf("$%d", seq)
}

type BindNameMssql struct {
	Seq
}

func (s *BindNameMssql) GetDialect() dialect.Dialect {
	return postgresql.NewPostgreSQLDialect()
}

func (s *BindNameMssql) NextBindName() string {
	s.seq++
	return fmt.Sprintf("@p%d", s.seq)
}

func (s *BindNameMssql) BindName(seq int) string {
	return fmt.Sprintf("@p%d", seq)
}

type BindNameMySQL struct {
	Seq
}

func (s *BindNameMySQL) GetDialect() dialect.Dialect {
	return mysql.NewMySQLDialect(mysql.SetANSIMode(true))
}

func (s *BindNameMySQL) NextBindName() string {
	s.seq++
	return "?"
}

func (s *BindNameMySQL) BindName(seq int) string {
	return "?"
}

func genDialect(driverName string) BindNameAware {
	switch driverName {
	case "oracle", "shentong":
		return &BindNameOracle{}
	case "postgres", "pgx", "kingbase":
		return &BindNamePg{}
	case "mssql", "sqlserver":
		return &BindNameMssql{}
	default:
		return &BindNameMySQL{}
	}
}

func GetBindPlaceholder(driverName string) func(seq int) string {
	switch driverName {
	case "oracle", "shentong":
		return func(seq int) string { return fmt.Sprintf(":%d", seq) }
	case "postgres", "pgx", "kingbase":
		return func(seq int) string { return fmt.Sprintf("$%d", seq) }
	case "mssql", "sqlserver":
		return func(seq int) string { return fmt.Sprintf("@p%d", seq) }
	default:
		return func(seq int) string { return "?" }
	}
}

type SQLVal struct {
	*sqlparser.SQLVal
}

// Format formats the node.
func (node *SQLVal) Format(buf *sqlparser.TrackedBuffer) {
	buf.WriteArg(string(node.Val))
}

type ColName struct {
	*sqlparser.ColName
	BindVar string
}

// Format formats the node.
func (node *ColName) Format(buf *sqlparser.TrackedBuffer) {
	buf.WriteArg(node.BindVar)
}

type FuncExpr struct {
	*sqlparser.FuncExpr
	BindVar string
}

// Format formats the node.
func (node *FuncExpr) Format(buf *sqlparser.TrackedBuffer) {
	buf.WriteArg(node.BindVar)
}

func fixBindPlaceholdersPrefix(q sqlparser.Statement, binder BindNameAware) (subEvals []ss.Subs) {
	return parseNodesForBindPlaceHolder(binder, q)
}

func parseNodesForBindPlaceHolder(binder BindNameAware, q ...sqlparser.SQLNode) (subEvals []ss.Subs) {
	visitor := newBindVarVisitor(binder)
	if err := sqlparser.Walk(visitor.visit, q...); err != nil {
		log.Printf("visitor error: %v", err)
		return
	}
	return visitor.finder.subEvals
}

func newBindVarVisitor(binder BindNameAware) *bindVarVisitor {
	return &bindVarVisitor{finder: createSubEvalFinder(binder)}
}

type bindVarVisitor struct {
	finder *subEvalFinder
}

func (v *bindVarVisitor) visit(node sqlparser.SQLNode) (kontinue bool, err error) {
	switch t := node.(type) {
	case *sqlparser.AliasedExpr:
		if val, ok := t.Expr.(*sqlparser.SQLVal); ok {
			t.Expr = v.finder.find(val)
		}
	case sqlparser.ValTuple:
		for j, expr := range t {
			t[j] = v.finder.find(expr)
		}
	case *sqlparser.UpdateExpr:
		t.Expr = v.finder.find(t.Expr)
	case *sqlparser.ComparisonExpr:
		t.Right = v.finder.find(t.Right)
		t.Left = v.finder.find(t.Left)
	}

	return true, nil
}

func createSubEvalFinder(binder BindNameAware) *subEvalFinder {
	return &subEvalFinder{binder: binder}
}

type subEvalFinder struct {
	binder   BindNameAware
	subEvals []ss.Subs
}

func (f *subEvalFinder) createBindPlaceHolder() string {
	return f.binder.NextBindName()
}

func (f *subEvalFinder) find(node sqlparser.Expr) sqlparser.Expr {
	switch t := node.(type) {
	case *sqlparser.ColName:
		if expr := sqlparser.String(t); strings.Contains(expr, "@") { // substitute function
			colName := &ColName{
				ColName: t,
				BindVar: f.createBindPlaceHolder(),
			}

			f.subEvals = append(f.subEvals, ss.ParseExpr(expr))
			return colName
		}
	case *sqlparser.FuncExpr:
		if name := t.Name.String(); strings.Contains(name, "@") { // substitute function
			funcExpr := &FuncExpr{
				FuncExpr: t,
				BindVar:  f.createBindPlaceHolder(),
			}

			f.subEvals = append(f.subEvals, ss.ParseExpr(sqlparser.String(t)))
			return funcExpr
		}
	case *sqlparser.SQLVal:
		if t.Type == sqlparser.StrVal && strings.Contains(string(t.Val), "@") {
			f.subEvals = append(f.subEvals, ss.ParseExpr(string(t.Val)))
			t.Val = []byte(f.createBindPlaceHolder())
			return &SQLVal{SQLVal: t}
		}
	}

	return node
}

// ParseSQL returns a normalized (lowercase's SQL commands) SQL string,
// and redacted SQL string with the params stripped out for display.
// Taken from sqlparser package
func ParseSQL(placeholder BindNameAware, sql string) (parsedQuery sqlparser.Statement, varSubs []ss.Subs, err error) {
	sqlStripped, _ := sqlparser.SplitMarginComments(sql)
	// sometimes queries might have ; at the end, that should be stripped
	sqlStripped = strings.TrimSuffix(sqlStripped, ";")
	stmt, err := Parse(sql, placeholder.GetDialect(), sqlparser.ModeStrict)
	if err == nil {
		return stmt, nil, nil
	}

	subs := ss.ParseExpr(sqlStripped)
	changedSQL := sqlStripped
	for _, sub := range subs {
		if t, ok := sub.(*ss.SubVar); ok {
			if old := `'` + t.Expr + `'`; strings.Contains(changedSQL, old) {
				changedSQL = strings.Replace(changedSQL, old, placeholder.NextBindName(), 1)
				varSubs = append(varSubs, ss.ParseExpr(t.Expr))
			}
		}
	}

	stmt, err = Parse(changedSQL, placeholder.GetDialect(), sqlparser.ModeDefault)
	return stmt, varSubs, err
}

// Parse using default dialect MySQl (for backward compatibility)
func Parse(sql string, dbDialect dialect.Dialect, mode sqlparser.Mode) (sqlparser.Statement, error) {
	statement, err := sqlparser.ParseWithDialect(dbDialect, sql)
	if err != nil && mode == sqlparser.ModeDefault {
		// log.Printf("ignoring error of non parsed sql statement")
		return sqlparser.NotParsedStatement{Query: sql}, nil
	}
	return statement, err
}
