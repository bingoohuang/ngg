package ui

import (
	"errors"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/AlecAivazis/survey/v2"
	"github.com/bingoohuang/ngg/godbtest/sqlmap"
	"github.com/bingoohuang/ngg/ss"
	"github.com/gohxs/readline"
)

type ItemAware interface {
	ItemTitle() string
	ItemDesc() string
}

func Select[T ItemAware](title string, items []T) (answerIndex int) {
	titles := make([]string, len(items))
	for i, m := range items {
		titles[i] = m.ItemTitle()
	}
	qs := &survey.Select{
		Message: title,
		Options: titles,
		Description: func(value string, index int) string {
			return items[index].ItemDesc()
		},
	}

	if err := survey.AskOne(qs, &answerIndex); err != nil {
		log.Print(err)
		os.Exit(1)
	}

	return answerIndex
}

type LineConfig struct {
	HistoryFile string
	Prefix      []string
	Suffix      []string
}

type LineConfigFn func(config *LineConfig)

func WithHistoryFile(value string) LineConfigFn {
	return func(c *LineConfig) {
		c.HistoryFile = value
	}
}

func WithPrefix(values ...string) LineConfigFn {
	return func(c *LineConfig) {
		c.Prefix = values
	}
}

func WithSuffix(values ...string) LineConfigFn {
	return func(c *LineConfig) {
		c.Suffix = values
	}
}

func GetSuffixes(sep string) []string {
	if sep == "" {
		return []string{
			";",  // 普通输出模式
			`\G`, // 纵向打印数据行（每行一个列值）
			`\j`, // JSON 打印数据行
			`\J`, // JSON 打印数据行
			`\I`, // Insert SQL 打印数据行
			`\P`, // Performance 性能压测模式
		}
	}

	return []string{
		sep,
		`\G`, // 纵向打印数据行（每行一个列值）
		`\j`, // JSON 打印数据行
		`\J`, // JSON 打印数据行
		`\I`, // Insert SQL 打印数据行
		`\P`, // Performance 性能压测模式
	}
}

func SepReg(sep string) *regexp.Regexp {
	reg := `(?i)`
	for i, suffix := range GetSuffixes(sep) {
		if i > 0 {
			reg += `|`
		}
		reg += `\` + suffix
	}
	// (?i);|\\G|\\J|\\I|\\P
	return regexp.MustCompile(reg)
}

func GetSQL(sep string) (string, error) {
	return ReadLine(
		WithHistoryFile(ss.Or(os.Getenv("HISTORY_FILE"), ".history")),
		WithPrefix(`%`, `@`),
		WithSuffix(GetSuffixes(sep)...))
}

var ErrExit = errors.New("exit")

func ReadLine(fns ...LineConfigFn) (string, error) {
	c := &LineConfig{
		HistoryFile: "/tmp/line",
	}
	for _, fn := range fns {
		fn(c)
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:                 "> ",
		HistoryFile:            c.HistoryFile,
		DisableAutoSaveHistory: true,
		Output:                 sqlmap.Color,
	})
	if err != nil {
		return "", err
	}

	defer ss.Close(rl)

	var lines []string
	var lastErr error
	var lastErrTime time.Time

	multipleLine := false

	for {
		line, err := rl.Readline()
		if err != nil {
			if errors.Is(err, readline.ErrInterrupt) {
				if errors.Is(lastErr, readline.ErrInterrupt) && time.Since(lastErrTime) <= time.Second {
					return "", ErrExit
				}

				lastErr = err
				lastErrTime = time.Now()
				lines = lines[:0]
				continue
			}
			break
		}
		if line = strings.TrimSpace(line); len(line) == 0 {
			continue
		}

		short := strings.TrimRightFunc(line, func(r rune) bool {
			return unicode.IsSpace(r) || r == ';'
		})
		if !multipleLine && ss.AnyOf(strings.ToLower(short), "exit", "quit") {
			return "", ErrExit
		}

		lines = append(lines, line)
		if ss.HasPrefix(line, c.Prefix...) {
			break
		}
		if !ss.HasSuffix(line, c.Suffix...) {
			rl.SetPrompt(">>> ")
			multipleLine = true
			continue
		}

		break
	}

	line := strings.Join(lines, "\n")
	_ = rl.SaveHistory(line)
	return line, nil
}
