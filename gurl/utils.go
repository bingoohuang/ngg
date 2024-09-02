package main

import (
	"hash"
	"io"
	"log"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/pflag"
)

func RemoveChars(input, cutset string) string {
	removeMap := func(r rune) rune {
		if strings.ContainsRune(cutset, r) {
			return -1 // 将 cutset 中的字符映射为 -1，表示移除
		}
		return r
	}

	// 使用 Map 函数应用映射函数到字符串中的每个字符
	return strings.Map(removeMap, input)
}

func HashFile(f string, h hash.Hash) ([]byte, error) {
	// 打开文件
	file, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err := io.Copy(h, file); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// TeeReadeCloser returns a ReadCloser that writes to w what it reads from r.
// All reads from r performed through it are matched with
// corresponding writes to w. There is no internal buffering -
// the write must complete before the read completes.
// Any error encountered while writing is reported as a read error.
func TeeReadeCloser(r io.ReadCloser, w io.Writer) io.ReadCloser {
	return &teeReadeCloser{r: r, w: w}
}

type teeReadeCloser struct {
	r io.ReadCloser
	w io.Writer
}

func (t *teeReadeCloser) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 {
		if n, err := t.w.Write(p[:n]); err != nil {
			return n, err
		}
	}
	return
}

func (t *teeReadeCloser) Close() error {
	return t.r.Close()
}

func surveyConfirm() {
	qs := []*survey.Question{{
		Name: "confirm",
		Prompt: &survey.Select{
			Message: "Please confirm your action:",
			Options: []string{"continue", "exit"},
			Default: "continue",
		},
	}}

	// the answers will be written to this struct
	answers := struct {
		Confirm string
	}{}

	// perform the questions
	if err := survey.Ask(qs, &answers); err != nil {
		log.Fatal(err)
	}

	if answers.Confirm == "exit" {
		os.Exit(0)
	}
}

func inSlice(str string, l []string) bool {
	for i := range l {
		if l[i] == str {
			return true
		}
	}
	return false
}

func flagEnvP(v *[]string, name, short, value, usage, envName string) {
	if value == "" {
		value = os.Getenv(envName)
	}
	var defaultValue []string
	if value != "" {
		defaultValue = []string{value}
	}

	pflag.StringSliceVarP(v, name, short, defaultValue, usage)
}

func flagEnvVar(p *string, name, value, usage, envName string) {
	if value == "" {
		value = os.Getenv(envName)
	}
	pflag.StringVar(p, name, value, usage)
}

func flagEnvVarP(p *string, name, short, value, usage, envName string) {
	if value == "" {
		value = os.Getenv(envName)
	}
	pflag.StringVarP(p, name, short, value, usage)
}
