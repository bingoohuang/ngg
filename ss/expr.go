package ss

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

type Valuer interface {
	Value(name, params, expr string) (any, error)
}

type ValuerHandler func(name, params string) (any, error)

func (f ValuerHandler) Value(name, params string) (any, error) { return f(name, params) }

func (s Subs) Eval(valuer Valuer) (any, error) {
	if len(s) == 1 && s.CountVars() == len(s) {
		v := s[0].(*SubVar)
		return valuer.Value(v.Name, v.Params, v.Expr)
	}

	value := ""
	for _, sub := range s {
		switch v := sub.(type) {
		case *SubTxt:
			value += v.Val
		case *SubVar:
			vv, err := valuer.Value(v.Name, v.Params, v.Expr)
			if err != nil {
				return nil, err
			}
			value += toString(vv)
		}
	}

	return value, nil
}

type SubTxt struct {
	Val string
}

func (s SubTxt) IsVar() bool { return false }

type SubVar struct {
	Name   string
	Params string
	Expr   string
}

func (s SubVar) IsVar() bool { return true }

type Sub interface {
	IsVar() bool
}

type Subs []Sub

func (s Subs) CountVars() (count int) {
	for _, sub := range s {
		if sub.IsVar() {
			count++
		}
	}

	return
}

func ParseExpr(src string) Subs {
	s := src
	var subs []Sub
	left := ""
	for {
		a := strings.Index(s, "@")
		if a < 0 || a == len(s)-1 {
			left += s
			break
		}

		left += s[:a]

		a++
		s = s[a:]
		if s[0] == '@' {
			s = s[1:]
			left += "@"
		} else if bracket := pairBracket(s[0]); bracket != nil {
			if rb := strings.IndexByte(s[1:], bracket.Right); rb > 0 {
				fn := s[1 : rb+1]
				s = s[rb+2:]

				subLiteral, subVar := parseName(&fn, &left, bracket)
				if subLiteral != nil {
					subs = append(subs, subLiteral)
				}
				if subVar != nil {
					subs = append(subs, subVar)
				}
			}
		} else {
			subLiteral, subVar := parseName(&s, &left, bracket)
			if subLiteral != nil {
				subs = append(subs, subLiteral)
			}
			if subVar != nil {
				subs = append(subs, subVar)
			}
		}
	}

	if left != "" {
		subs = append(subs, &SubTxt{Val: left})
	}

	if Subs(subs).CountVars() == 0 {
		return []Sub{&SubTxt{Val: src}}
	}

	return subs
}

type bracket struct {
	Left  byte
	Right byte
}

func pairBracket(left byte) *bracket {
	switch left {
	case '{':
		return &bracket{Left: '{', Right: '}'}
	case '[':
		return &bracket{Left: '[', Right: ']'}
	case '#', '%', '`':
		return &bracket{Left: left, Right: left}
	case '<':
		return &bracket{Left: '<', Right: '>'}
	}
	return nil
}

func parseName(s, left *string, bracket *bracket) (subLiteral, subVar Sub) {
	original := *s
	name := ""
	offset := 0
	for i, r := range *s {
		if !validNameRune(r) {
			name = (*s)[:i]
			break
		}
		offset += utf8.RuneLen(r)
	}

	nonParam := name == "" && offset == len(*s)
	if nonParam {
		name = *s
	}

	if *left != "" {
		subLiteral = &SubTxt{Val: *left}
		*left = ""
	}

	sv := &SubVar{
		Name: name,
	}
	subVar = sv

	if !nonParam && offset > 0 && offset < len(*s) {
		if (*s)[offset] == '(' {
			if rb := strings.IndexByte(*s, ')'); rb > 0 {
				sv.Params = (*s)[offset+1 : rb]
				*s = (*s)[rb+1:]
				sv.Expr = wrap(original[:rb+1], bracket)
				return
			}
		}
	}

	*s = (*s)[offset:]
	sv.Expr = wrap(original[:offset], bracket)

	return
}

func wrap(s string, bracket *bracket) string {
	if bracket != nil {
		return "@" + string(bracket.Left) + s + string(bracket.Right)
	}

	return "@" + s
}

func validNameRune(r int32) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Han, r) ||
		r == '_' || r == '-' || r == '.'
}

func toString(value any) string {
	switch vv := value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", vv)
	case float32, float64:
		return fmt.Sprintf("%f", vv)
	case bool:
		return fmt.Sprintf("%t", vv)
	case string:
		return vv
	default:
		vvv := fmt.Sprintf("%v", value)
		return vvv
	}
}

type GenFn func() any

type GenFnFn func(params string) GenFn

type MapGenValue struct {
	Map map[string]GenFn
	sync.RWMutex

	GenMap     map[string]GenFnFn
	MissedVars map[string]bool
	Vars       map[string]any
}

func NewMapGenValue(m map[string]GenFnFn) *MapGenValue {
	return &MapGenValue{
		GenMap:     m,
		Map:        map[string]GenFn{},
		Vars:       map[string]any{},
		MissedVars: map[string]bool{},
	}
}

func (m *MapGenValue) Value(name, params, expr string) any {
	return m.GetValue(name, params, expr)
}

func (m *MapGenValue) GetValue(name, params, expr string) any {
	m.RLock()
	if fn, ok := m.Map[name]; ok {
		m.RUnlock()
		return fn()
	}
	m.RUnlock()

	var f GenFn

	m.Lock()
	defer m.Unlock()

	if fn, ok := m.GenMap[name]; ok {
		ff := fn(params)
		f = func() any {
			v := ff()
			m.Vars[name] = v
			return v
		}
	} else {
		f = func() any { return expr }
		m.MissedVars[name] = true
	}

	m.Map[name] = f
	return f()
}

type VarValue interface {
	GetValue(name, params, expr string) any
}

type VarValueHandler func(name, params, expr string) any

func (v VarValueHandler) GetValue(name, params, expr string) any {
	return v(name, params, expr)
}

func Eval(s string, varValue VarValue) string {
	return ParseSubstitute(s).Eval(varValue)
}

type Part interface {
	Eval(varValue VarValue) string
}

type Var struct {
	Name string
	Expr string
}

type Literal struct{ V string }

func (l Literal) Eval(VarValue) string { return l.V }
func (l Var) Eval(varValue VarValue) string {
	return fmt.Sprintf("%s", varValue.GetValue(l.Name, "", l.Expr))
}

func (l Parts) Eval(varValue VarValue) string {
	sb := strings.Builder{}
	for _, p := range l {
		sb.WriteString(p.Eval(varValue))
	}
	return sb.String()
}

type Parts []Part

var varRe = regexp.MustCompile(`\$?\{[^{}]+?\}|\{\{[^{}]+?\}\}`)

func ParseSubstitute(s string) (parts Parts) {
	locs := varRe.FindAllStringSubmatchIndex(s, -1)
	start := 0

	for _, loc := range locs {
		parts = append(parts, &Literal{V: s[start:loc[0]]})
		sub := s[loc[0]+1 : loc[1]-1]
		sub = strings.TrimPrefix(sub, "{")
		sub = strings.TrimSuffix(sub, "}")
		start = loc[1]

		vn := strings.TrimSpace(sub)

		parts = append(parts, &Var{Name: vn, Expr: sub})
	}

	if start < len(s) {
		parts = append(parts, &Literal{V: s[start:]})
	}

	return parts
}
