package gterm

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/bingoohuang/ngg/gum"
	"github.com/bingoohuang/ngg/ss"
)

type Option struct {
	RandomSize  int64
	Random      bool
	Required    bool
	StripSpaces bool
	TryDecode   bool
}

type ReaderSource int

const (
	SourceNone ReaderSource = iota
	SourcePrompt
	SourceRandom
	SourceFile
	SourceInput
	SourcePipe
)

type Result struct {
	io.Reader
	Len         int
	SourceType  ReaderSource
	SourceTitle string
	Bytes       []byte
}

func (r Result) ToBytes() ([]byte, error) {
	if len(r.Bytes) > 0 {
		return r.Bytes, nil
	}

	return io.ReadAll(r.Reader)
}

func (o Option) Open(input string) (*Result, error) {
	if o.Random && o.RandomSize <= 0 {
		o.RandomSize = 64
	}

	if input == "" {
		// https://www.socketloop.com/tutorials/golang-check-if-os-stdin-input-data-is-piped-or-from-terminal
		fi, _ := os.Stdin.Stat() // get the FileInfo struct describing the standard input.

		if (fi.Mode() & os.ModeCharDevice) == 0 {
			// do things for data from pipe
			dat, err := io.ReadAll(os.Stdin)
			if err != nil {
				return nil, err
			}

			r := &Result{
				Reader:      bytes.NewReader(dat),
				Bytes:       dat,
				Len:         len(dat),
				SourceType:  SourcePipe,
				SourceTitle: "pipe",
			}
			return r, nil

		}

		if o.Required {
			line, err := gum.Input("Input:", "Input something")
			if err != nil {
				return nil, err
			}

			format := SplitTail(&line, ':')

			return &Result{
				Reader:      decorateReader(strings.NewReader(line), format),
				Bytes:       []byte(line),
				Len:         len(line),
				SourceType:  SourcePrompt,
				SourceTitle: "prompt",
			}, nil

		}

		if o.Random {
			return &Result{
				Reader:      io.LimitReader(rand.Reader, o.RandomSize),
				Len:         int(o.RandomSize),
				SourceType:  SourceRandom,
				SourceTitle: "random",
			}, nil
		}
	}

	format := SplitTail(&input, ':')

	if stat, err := os.Stat(input); err == nil && !stat.IsDir() {
		f, err := os.Open(input)
		if err != nil {
			return nil, err
		}

		return &Result{
			Reader:      decorateReader(f, format),
			Len:         int(stat.Size()),
			SourceType:  SourceFile,
			SourceTitle: "file " + input,
		}, nil
	}

	name := ss.Abbreviate(input, 10, ss.DefaultEllipse)
	if o.StripSpaces {
		input = StripSpaces(input)
	}

	return &Result{
		Reader:      decorateReader(strings.NewReader(input), format),
		Len:         len(input),
		SourceType:  SourceInput,
		SourceTitle: name,
	}, nil
}

func SplitTail(s *string, c byte) (tail string) {
	if p := strings.LastIndexByte(*s, c); p >= 0 {
		tail = (*s)[p+1:]
		*s = (*s)[:p]
	}
	return tail
}

func DecodeByTailTag(s string) ([]byte, error) {
	format := SplitTail(&s, ':')
	switch format {
	case "hex":
		return hex.DecodeString(s)
	case "base64", "b64":
		return base64.StdEncoding.DecodeString(s)
	default:
		return []byte(s), nil
	}
}

func decorateReader(r io.Reader, format string) io.Reader {
	switch strings.ToLower(format) {
	case "hex":
		return hex.NewDecoder(r)
	case "base64", "b64":
		return base64.NewDecoder(base64.StdEncoding, r)
	default:
		return r
	}
}

func ParsePrefix(value string) (prefix string, data []byte, err error) {
	if strings.HasPrefix(value, "raw:") {
		return "raw:", []byte(value[len("raw:"):]), nil
	}
	if ss.HasPrefix(value, "base64:") {
		d, err := ss.Base64().DecodeBytes([]byte(value[len("base64:"):]))
		return "base64:", d.Bytes(), err
	}
	if ss.HasPrefix(value, "b64:") {
		d, err := ss.Base64().DecodeBytes([]byte(value[len("b64:"):]))
		return "base64:", d.Bytes(), err
	}

	if strings.HasPrefix(value, "hex:") {
		data, err = hex.DecodeString(value[len("hex:"):])
		return "hex:", data, err
	}

	return "", []byte(value), nil
}

func StripSpaces(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1 // if the character is a space, drop it
		}

		return r // else keep it in the string
	}, str)
}

func (r Result) Close() {
	if c, ok := r.Reader.(io.Closer); ok {
		if err := c.Close(); err != nil {
			log.Printf("close error: %v", err)
		}
	}
}
