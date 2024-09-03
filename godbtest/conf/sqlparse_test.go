package conf

import (
	"bytes"
	"testing"

	"github.com/bingoohuang/ngg/pp"
	"github.com/bingoohuang/ngg/sqlparser"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestParseSelectFunc(t *testing.T) {
	dialectFn := genDialect("oracle")
	query := `SELECT F_1('@uuid','21',TO_DATE('2014-07-02', 'YYYY-MM-DD'),TO_DATE('2014-07-02', 'YYYY-MM-DD'),256,3,100) AS VALUE FROM DUAL`
	parsedQuery, subEvals, err := ParseSQL(dialectFn, query)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(subEvals))
	subEvals = fixBindPlaceholdersPrefix(parsedQuery, dialectFn)
	fullSQL := sqlparser.String(parsedQuery)
	assert.Equal(t, `select F_1(:1, '21', TO_DATE('2014-07-02', 'YYYY-MM-DD'), TO_DATE('2014-07-02', 'YYYY-MM-DD'), 256, 3, 100) as VALUE from dual`, fullSQL)
	assert.Equal(t, 1, len(subEvals))
}

func TestInsert(t *testing.T) {
	dialectFn := genDialect("mysql")
	query := `insert into t2 values('@ksuid','@姓名','@地址','@邮箱','@手机','@random_int(15-95)','@身份证')`
	parsedQuery, subEvals, err := ParseSQL(dialectFn, query)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(subEvals))
	subEvals = fixBindPlaceholdersPrefix(parsedQuery, dialectFn)
	fullSQL := sqlparser.String(parsedQuery)
	assert.Equal(t, `insert into t2 values (?, ?, ?, ?, ?, ?, ?)`, fullSQL)
	assert.Equal(t, 7, len(subEvals))
}

func TestInsert2(t *testing.T) {
	dialectFn := genDialect("oracle")
	query := `insert into KM(KEYID, KMKEYID, KMID, KEYALG, KEYLEN, KEYFLAG, PKEYHASH, GENDATE, KEYSTAT, BACKUPFLAG) ` +
		`values (SEQ_KM.NEXTVAL, 17090451, '2', 3, 256, 7, 'BCxdqNG7rbA=', TO_TIMESTAMP('2014-07-02 06:14:00.742000000', 'YYYY-MM-DD HH24:MI:SS.FF'), 0, '0')`
	parsedQuery, subEvals, err := ParseSQL(dialectFn, query)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(subEvals))
	subEvals = fixBindPlaceholdersPrefix(parsedQuery, dialectFn)
	fullSQL := sqlparser.String(parsedQuery)
	assert.Equal(t, query, fullSQL)
	assert.Equal(t, 0, len(subEvals))
}

// TODO: oracle SQL 中 current_timestamp， 解析结构再次生成SQL时，变换成了 current_timestamp()
func TestInsert3(t *testing.T) {
	dialectFn := genDialect("oracle")
	query := `insert into KM(KEYID, KMKEYID, KMID, KEYALG, KEYLEN, KEYFLAG, PKEYHASH, GENDATE, KEYSTAT, BACKUPFLAG) ` +
		`values (SEQ_KM_KEY_DATA.NEXTVAL, 17090451, '2', 3, 256, 7, 'BCvHn', current_timestamp, 0, '0')`
	parsedQuery, subEvals, err := ParseSQL(dialectFn, query)

	query1 := `insert into KM(KEYID, KMKEYID, KMID, KEYALG, KEYLEN, KEYFLAG, PKEYHASH, GENDATE, KEYSTAT, BACKUPFLAG) ` +
		`values (SEQ_KM_KEY_DATA.NEXTVAL, 17090451, '2', 3, 256, 7, 'BCvHn', current_timestamp(), 0, '0')`
	parsedQuery1, subEvals, err := ParseSQL(dialectFn, query1)

	var b1, b2 bytes.Buffer
	pn := pp.New().SetOmitEmpty(true).SetColoring(false)
	pn.Fprint(&b1, parsedQuery)
	pn.Fprint(&b2, parsedQuery1)
	t.Log("b1: ", b1.String())
	t.Log("b2: ", b2.String())
	t.Log("Diff: ", cmp.Diff(b1.String(), b2.String()))

	assert.Nil(t, err)
	assert.Equal(t, 0, len(subEvals))
	subEvals = fixBindPlaceholdersPrefix(parsedQuery, dialectFn)
	fullSQL := sqlparser.String(parsedQuery)
	assert.Equal(t, query1, fullSQL)
	assert.Equal(t, 0, len(subEvals))
}

func TestQuery(t *testing.T) {
	dialectFn := genDialect("mysql")
	query := `select * from t2 where id = '@ksuid'`
	parsedQuery, subEvals, err := ParseSQL(dialectFn, query)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(subEvals))
	subEvals = fixBindPlaceholdersPrefix(parsedQuery, dialectFn)
	fullSQL := sqlparser.String(parsedQuery)
	assert.Equal(t, `select * from t2 where id = ?`, fullSQL)
	assert.Equal(t, 1, len(subEvals))
}

func TestUpdate(t *testing.T) {
	dialectFn := genDialect("mysql")
	query := `update t2 set name = '@姓名' where id = '@ksuid'`
	parsedQuery, subEvals, err := ParseSQL(dialectFn, query)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(subEvals))
	subEvals = fixBindPlaceholdersPrefix(parsedQuery, dialectFn)
	fullSQL := sqlparser.String(parsedQuery)
	assert.Equal(t, `update t2 set name = ? where id = ?`, fullSQL)
	assert.Equal(t, 2, len(subEvals))
}
