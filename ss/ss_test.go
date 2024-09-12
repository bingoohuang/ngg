package ss_test

import (
	"fmt"
	"testing"

	"github.com/bingoohuang/ngg/ss"
	"github.com/stretchr/testify/assert"
)

func ExampleIf() {
	fmt.Println(ss.If(true, "bingoo", "huang"))
	fmt.Println(ss.If(false, "bingoo", "huang"))
	// Output:
	// bingoo
	// huang
}

func TestMapGet(t *testing.T) {
	m := map[string]string{"a": "b"}
	assert.Equal(t, "b", ss.MapGet(m, "a", ""))
	assert.Equal(t, "c", ss.MapGetF(m, "b", func() string { return "c" }))
	assert.Equal(t, "d", ss.MapGet(m, "b", "d"))
}

func TestSplit(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, ss.Split(" a ,b , ,", ","))

	assert.Equal(t, []string{"a", "b"}, ss.SplitSeps(",a;b,", ",;", -1))

}

func TestAbbreviate(t *testing.T) {
	s := ss.Abbreviate(`内蒙古自治区乌兰察布市嗪貶路5924号黢蹵小区11单元1002室`, 16, "…")
	assert.Equal(t, `内蒙古自治区乌兰察布市嗪貶路59…`, s)
}

var tests = []struct {
	raw    string
	quoted string
}{
	{"", "''"},
	{"hello", "'hello'"},
	{"hello world", "'hello world'"},
	{`rock'n'roll`, `'rock\'n\'roll'`},
	{`"rock'n'roll"`, `'"rock\'n\'roll"'`},
	{`rock\\'n\\'roll`, `'rock\\\'n\\\'roll'`},
}

func TestQuote(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got := ss.QuoteSingle(tt.raw)
			if got != tt.quoted {
				t.Errorf("got %v\nwant %v", got, tt.quoted)
			}
		})
	}
}

func TestUnquote(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.quoted, func(t *testing.T) {
			got, err := ss.UnquoteSingle(tt.quoted)
			if err != nil {
				t.Error(err)
			}
			if got != tt.raw {
				t.Errorf("got %v\nwant %v", got, tt.raw)
			}
		})
	}
}

func TestUnquoteError(t *testing.T) {
	tests := []struct {
		quoted  string
		wantErr bool
	}{
		{"hello world", true},
		{"'hello world", true},
		{"hello world'", true},
		{`''hello'`, true},
		{`'\'hello'`, false},
		{`'\\'hello'`, true},
		{`''`, false},
		{`h`, true},
	}
	for _, tt := range tests {
		t.Run(tt.quoted, func(t *testing.T) {
			_, err := ss.UnquoteSingle(tt.quoted)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Error(err)
			}
			if tt.wantErr {
				t.Error("want error")
			}
		})
	}
}

func ExampleQuoteSingle() {
	s := ss.QuoteSingle(`"Fran & Freddie's Diner	☺"`)
	fmt.Println(s)
	s = ss.QuoteSingle(`rock'n'roll`)
	fmt.Println(s)
	// Output:
	// '"Fran & Freddie\'s Diner	☺"'
	// 'rock\'n\'roll'
}

func ExampleUnquoteSingle() {
	s, err := ss.UnquoteSingle("You can't unquote a string without quotes")
	fmt.Printf("%q, %v\n", s, err)
	s, err = ss.UnquoteSingle("\"The string must be either double-quoted\"")
	fmt.Printf("%q, %v\n", s, err)
	s, err = ss.UnquoteSingle("`or backquoted.`")
	fmt.Printf("%q, %v\n", s, err)
	s, err = ss.UnquoteSingle("'\u263a'")
	fmt.Printf("%q, %v\n", s, err)
	s, err = ss.UnquoteSingle("'\u2639\u2639'")
	fmt.Printf("%q, %v\n", s, err)
	s, err = ss.UnquoteSingle("'\\'The string must be either single-quoted\\''")
	fmt.Printf("%q, %v\n", s, err)
	// Output:
	// "", invalid syntax
	// "", invalid syntax
	// "", invalid syntax
	// "☺", <nil>
	// "☹☹", <nil>
	// "'The string must be either single-quoted'", <nil>
}
