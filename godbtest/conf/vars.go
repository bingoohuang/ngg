package conf

import (
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/bingoohuang/ngg/godbtest/ui"
	"github.com/bingoohuang/ngg/ss"
)

func (c *SubVars) evalSQL(sep string) []string {
	// the questions to ask
	var qs []*survey.Question

	for name := range c.varNames {
		qs = append(qs, &survey.Question{
			Name:     name,
			Validate: survey.Required,
			Prompt:   &survey.Input{Message: name},
		})
	}

	answers := newMapAnswers()
	// perform the questions
	if err := survey.Ask(qs, answers); err != nil {
		log.Print(err)
		os.Exit(1)
	}

	evalSql := c.parts.Eval(ss.VarValueHandler(func(name, params, expr string) any {
		return answers.answers[name]
	}))

	return ss.SplitReg(evalSql, ui.SepReg(sep), -1)
}

type SubVars struct {
	varNames map[string]bool
	parts    ss.Parts
}

func parseVars(s string) *SubVars {
	parts := ParseSubstitute(s)
	varNames := make(map[string]bool)

	for _, p := range parts {
		if va, ok := p.(*ss.Var); ok {
			varNames[va.Name] = true
		}
	}

	if len(varNames) > 0 {
		return &SubVars{
			varNames: varNames,
			parts:    parts,
		}
	}

	return nil
}

func (c *Config) parserVars() {
	for _, a := range c.Actions {
		for _, s := range a.Sqls {
			s.subVars = parseVars(s.Sql)
		}
	}
}

type mapAnswers struct {
	answers map[string]any
}

func newMapAnswers() *mapAnswers {
	return &mapAnswers{
		answers: make(map[string]any),
	}
}

func (m *mapAnswers) WriteAnswer(field string, value any) error {
	m.answers[field] = value
	return nil
}

var _ core.Settable = (*mapAnswers)(nil)

var varRe = regexp.MustCompile(`\{\{[^{}]+?}}`)

func ParseSubstitute(s string) (parts ss.Parts) {
	locs := varRe.FindAllStringSubmatchIndex(s, -1)
	start := 0

	for _, loc := range locs {
		parts = append(parts, &ss.Literal{V: s[start:loc[0]]})
		sub := s[loc[0]+1 : loc[1]-1]
		sub = strings.TrimPrefix(sub, "{")
		sub = strings.TrimSuffix(sub, "}")
		start = loc[1]

		vn := strings.TrimSpace(sub)

		parts = append(parts, &ss.Var{Name: vn, Expr: sub})
	}

	if start < len(s) {
		parts = append(parts, &ss.Literal{V: s[start:]})
	}

	return parts
}
