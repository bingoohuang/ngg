package label

import (
	"errors"
	"go/ast"
	"go/parser"
	"strconv"
)

const (
	idLogic = iota
	idOp
	idName
	idValue
	idEvald

	opAnd = 34
	opOr  = 35
)

var (
	ErrNotSupportOperator = errors.New("operator doesn't support")
	ErrQuery              = errors.New("query error")
)

type Visitor struct {
	idents []ident
}

type ident struct {
	value string
	vType int
}

func (v *Visitor) Visit(n ast.Node) ast.Visitor {
	switch d := n.(type) {
	case *ast.Ident:
		v.idents = append(v.idents, ident{d.Name, idName})
	case *ast.BasicLit:
		v.idents = append(v.idents, ident{d.Value, idValue})
	case *ast.BinaryExpr:
		if d.Op == opAnd || d.Op == opOr {
			v.idents = append(v.idents, ident{d.Op.String(), idLogic})
		} else {
			v.idents = append(v.idents, ident{d.Op.String(), idOp})
		}

	}

	return v
}

func Parse(expr string) (*Visitor, error) {
	e, err := parser.ParseExpr(expr)
	if err != nil {
		return nil, err
	}

	var v Visitor
	ast.Walk(&v, e)

	return &v, nil
}

func (v *Visitor) Eval(labels map[string]string) (bool, error) {
	var results []ident
	i := 0

	for {
		if len(v.idents[i:]) < 3 {
			break
		}

		if v.idents[i].vType == 0 {
			results = append(results, v.idents[i])
			i++
			continue
		}

		if relOpEval(v.idents[i:i+3], labels) {
			results = append(results, ident{"true", idEvald})
		} else {
			results = append(results, ident{"false", idEvald})
		}

		if i+3 < len(v.idents) {
			i += 3
		} else {
			break
		}
	}

	if len(results) < 1 {
		return false, ErrQuery
	}

	ok, err := binOpEval(&results)
	if err != nil {
		return ok, err
	}

	return ok, nil
}

func binOpEval(idents *[]ident) (bool, error) {
	i := 0
	for {
		if len(*idents) == 1 {
			r, _ := strconv.ParseBool((*idents)[0].value)
			return r, nil
		}

		if (*idents)[i].vType == 0 && (*idents)[i+1].vType == 0 {
			i++
			continue
		}

		if len(*idents) > 2 && (*idents)[i].vType == idLogic &&
			(*idents)[i+1].vType == idEvald && (*idents)[i+2].vType == idEvald {

			r1, _ := strconv.ParseBool((*idents)[i+1].value)
			r2, _ := strconv.ParseBool((*idents)[i+2].value)

			switch {
			case (*idents)[i].value == "&&":
				(*idents)[i].value = strconv.FormatBool(r1 && r2)
			case (*idents)[i+0].value == "||":
				(*idents)[i].value = strconv.FormatBool(r1 || r2)
			default:
				return false, ErrNotSupportOperator
			}

			(*idents)[i].vType = idEvald
			if len(*idents) > 2 {
				*idents = append((*idents)[:i+1], (*idents)[i+3:]...)
				i = 0
				continue
			} else {
				tmp := (*idents)[0:1]
				*idents = tmp
			}
		}

		if len(*idents) > 4 {
			i += 2
		} else {
			return false, ErrNotSupportOperator
		}
	}
}

func relOpEval(idents []ident, labels map[string]string) bool {
	if idents[0].value == "==" {
		if _, ok := labels[idents[1].value]; ok {
			return labels[idents[1].value] == idents[2].value
		}
	} else {
		if _, ok := labels[idents[1].value]; ok {
			return labels[idents[1].value] != idents[2].value
		}
	}

	return false
}
