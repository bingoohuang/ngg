/*
Copyright 2017 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sqlparser

import (
	"fmt"
	"github.com/bingoohuang/ngg/sqlparser/dialect"
	"github.com/bingoohuang/ngg/sqlparser/dialect/mysql"
	"github.com/bingoohuang/ngg/sqlparser/dialect/postgresql"
	"reflect"
	"testing"
)

func TestLiteralID(t *testing.T) {
	testcases := []struct {
		in  string
		id  int
		out string
	}{{
		in:  "`aa`",
		id:  ID,
		out: "aa",
	}, {
		in:  "```a```",
		id:  ID,
		out: "`a`",
	}, {
		in:  "`a``b`",
		id:  ID,
		out: "a`b",
	}, {
		in:  "`a``b`c",
		id:  ID,
		out: "a`b",
	}, {
		in:  "`a``b",
		id:  LEX_ERROR,
		out: "a`b",
	}, {
		in:  "`a``b``",
		id:  LEX_ERROR,
		out: "a`b`",
	}, {
		in:  "``",
		id:  LEX_ERROR,
		out: "",
	}}

	for _, tcase := range testcases {
		tkn := NewStringTokenizer(tcase.in)
		id, out := tkn.Scan()
		if tcase.id != id || string(out) != tcase.out {
			t.Errorf("Scan(%s): %d, %s, want %d, %s", tcase.in, id, out, tcase.id, tcase.out)
		}
	}
}

func tokenName(id int) string {
	switch id {
	case SINGLE_QUOTE_STRING:
		return "SINGLE_QUOTE_STRING"
	case DOUBLE_QUOTE_STRING:
		return "DOUBLE_QUOTE_STRING"
	case BACK_QUOTE_STRING:
		return "BACK_QUOTE_STRING"
	case LEX_ERROR:
		return "LEX_ERROR"
	}
	return fmt.Sprintf("%d", id)
}

func TestString(t *testing.T) {
	testcases := []struct {
		in   string
		id   int
		want string
	}{{
		in:   "''",
		id:   SINGLE_QUOTE_STRING,
		want: "",
	}, {
		in:   "''''",
		id:   SINGLE_QUOTE_STRING,
		want: "'",
	}, {
		in:   "'hello'",
		id:   SINGLE_QUOTE_STRING,
		want: "hello",
	}, {
		in:   "'\\n'",
		id:   SINGLE_QUOTE_STRING,
		want: "\n",
	}, {
		in:   "'\\nhello\\n'",
		id:   SINGLE_QUOTE_STRING,
		want: "\nhello\n",
	}, {
		in:   "'a''b'",
		id:   SINGLE_QUOTE_STRING,
		want: "a'b",
	}, {
		in:   "'a\\'b'",
		id:   SINGLE_QUOTE_STRING,
		want: "a'b",
	}, {
		in:   "'\\'",
		id:   LEX_ERROR,
		want: "'",
	}, {
		in:   "'",
		id:   LEX_ERROR,
		want: "",
	}, {
		in:   "'hello\\'",
		id:   LEX_ERROR,
		want: "hello'",
	}, {
		in:   "'hello",
		id:   LEX_ERROR,
		want: "hello",
	}, {
		in:   "'hello\\",
		id:   LEX_ERROR,
		want: "hello",
	}}

	for _, tcase := range testcases {
		id, got := NewStringTokenizer(tcase.in).Scan()
		if tcase.id != id || string(got) != tcase.want {
			t.Errorf("Scan(%q) = (%s, %q), want (%s, %q)", tcase.in, tokenName(id), got, tokenName(tcase.id), tcase.want)
		}
	}
}

func TestSplitStatement(t *testing.T) {
	testcases := []struct {
		in  string
		sql string
		rem string
	}{{
		in:  "select * from table",
		sql: "select * from table",
	}, {
		in:  "select * from table; ",
		sql: "select * from table",
		rem: " ",
	}, {
		in:  "select * from table; select * from table2;",
		sql: "select * from table",
		rem: " select * from table2;",
	}, {
		in:  "select * from /* comment */ table;",
		sql: "select * from /* comment */ table",
	}, {
		in:  "select * from /* comment ; */ table;",
		sql: "select * from /* comment ; */ table",
	}, {
		in:  "select * from table where semi = ';';",
		sql: "select * from table where semi = ';'",
	}, {
		in:  "-- select * from table",
		sql: "-- select * from table",
	}, {
		in:  " ",
		sql: " ",
	}, {
		in:  "",
		sql: "",
	}}

	for _, tcase := range testcases {
		sql, rem, err := SplitStatement(tcase.in)
		if err != nil {
			t.Errorf("EndOfStatementPosition(%s): ERROR: %v", tcase.in, err)
			continue
		}

		if tcase.sql != sql {
			t.Errorf("EndOfStatementPosition(%s) got sql \"%s\" want \"%s\"", tcase.in, sql, tcase.sql)
		}

		if tcase.rem != rem {
			t.Errorf("EndOfStatementPosition(%s) got remainder \"%s\" want \"%s\"", tcase.in, rem, tcase.rem)
		}
	}
}

func TestSetDefaultDialect(t *testing.T) {
	tests := []dialect.Dialect{
		mysql.NewMySQLDialect(),
		mysql.NewMySQLDialect(mysql.SetANSIMode(true)),
		postgresql.NewPostgreSQLDialect(),
	}
	for _, testDialect := range tests {
		t.Run("check SetDefaultDialect that it changes global defaultDialect", func(t *testing.T) {
			SetDefaultDialect(testDialect)
			if reflect.TypeOf(testDialect) != reflect.TypeOf(defaultDialect) {
				t.Fatal("incorrectly set default dialect")
			}
		})
	}
}