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
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/bingoohuang/ngg/sqlparser/dialect"
	"github.com/bingoohuang/ngg/sqlparser/dialect/mysql"
)

// TestParseNextValid concatenates all the valid SQL test cases and check it can read
// them as one long string.
func TestParseNextValid(t *testing.T) {
	var sql bytes.Buffer
	for _, tcase := range validSQL {
		sql.WriteString(strings.TrimSuffix(tcase.input, ";"))
		sql.WriteRune(';')
	}

	var dialect dialect.Dialect
	tokenizer := NewTokenizer(&sql)
	for i, tcase := range validSQL {
		dialect = tcase.dialect
		if dialect == nil {
			dialect = mysql.NewMySQLDialect()
		}
		tokenizer.dialect = dialect
		input := tcase.input + ";"
		want := tcase.output
		if want == "" {
			want = tcase.input
		}

		tree, err := ParseNext(tokenizer)
		if err != nil {
			t.Fatalf("[%d] ParseNext(%q) err: %q, want nil", i, input, err)
		}

		if got := StringWithDialect(dialect, tree); got != want {
			t.Fatalf("[%d] ParseNext(%q) = %q, want %q", i, input, got, want)
		}
	}

	// Read once more and it should be EOF.
	if tree, err := ParseNext(tokenizer); err != io.EOF {
		t.Errorf("ParseNext(tokenizer) = (%q, %v) want io.EOF", String(tree), err)
	}
}

// TestParseNextErrors tests all the error cases, and ensures a valid
// SQL statement can be passed afterwards.
func TestParseNextErrors(t *testing.T) {
	var testDialect dialect.Dialect
	SetTokenizerVerbosity(true)
	for _, tcase := range invalidSQL {
		if tcase.excludeMulti {
			// Skip tests which leave unclosed strings, or comments.
			continue
		}

		testDialect = tcase.dialect
		if testDialect == nil {
			testDialect = mysql.NewMySQLDialect()
		}

		sql := tcase.input + "; select 1 from t"
		tokens := NewStringTokenizerWithDialect(testDialect, sql)

		// The first statement should be an error
		_, err := ParseNext(tokens)
		if err == nil || err.Error() != tcase.output {
			t.Fatalf("[0] ParseNext(%q) err: %q, want %q", sql, err, tcase.output)
		}

		// The second should be valid
		tree, err := ParseNext(tokens)
		if err != nil {
			t.Fatalf("[1] ParseNext(%q) err: %q, want nil", sql, err)
		}

		want := "select 1 from t"
		if got := String(tree); got != want {
			t.Fatalf("[1] ParseNext(%q) = %q, want %q", sql, got, want)
		}

		// Read once more and it should be EOF.
		if tree, err := ParseNext(tokens); err != io.EOF {
			t.Errorf("ParseNext(tokens) = (%q, %v) want io.EOF", String(tree), err)
		}
	}
}

// TestParseNextEdgeCases tests various ParseNext edge cases.
func TestParseNextEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{{
		name:  "Trailing ;",
		input: "select 1 from a; update a set b = 2;",
		want:  []string{"select 1 from a", "update a set b = 2"},
	}, {
		name:  "No trailing ;",
		input: "select 1 from a; update a set b = 2",
		want:  []string{"select 1 from a", "update a set b = 2"},
	}, {
		name:  "Trailing whitespace",
		input: "select 1 from a; update a set b = 2    ",
		want:  []string{"select 1 from a", "update a set b = 2"},
	}, {
		name:  "Trailing whitespace and ;",
		input: "select 1 from a; update a set b = 2   ;   ",
		want:  []string{"select 1 from a", "update a set b = 2"},
	}, {
		name:  "Handle ForceEOF statements",
		input: "set character set utf8; select 1 from a",
		want:  []string{"set charset 'utf8'", "select 1 from a"},
	}, {
		name:  "Semicolin inside a string",
		input: "set character set ';'; select 1 from a",
		want:  []string{"set charset ';'", "select 1 from a"},
	}, {
		name:  "Partial DDL",
		input: "create table a; select 1 from a",
		want:  []string{"create table a", "select 1 from a"},
	}, {
		name:  "Partial DDL",
		input: "create table a ignore me this is garbage; select 1 from a",
		want:  []string{"create table a", "select 1 from a"},
	}}

	for _, test := range tests {
		tokens := NewStringTokenizer(test.input)

		for i, want := range test.want {
			tree, err := ParseNext(tokens)
			if err != nil {
				t.Fatalf("[%d] ParseNext(%q) err = %q, want nil", i, test.input, err)
			}

			if got := String(tree); got != want {
				t.Fatalf("[%d] ParseNext(%q) = %q, want %q", i, test.input, got, want)
			}
		}

		// Read once more and it should be EOF.
		if tree, err := ParseNext(tokens); err != io.EOF {
			t.Errorf("ParseNext(%q) = (%q, %v) want io.EOF", test.input, String(tree), err)
		}

		// And again, once more should be EOF.
		if tree, err := ParseNext(tokens); err != io.EOF {
			t.Errorf("ParseNext(%q) = (%q, %v) want io.EOF", test.input, String(tree), err)
		}
	}
}
