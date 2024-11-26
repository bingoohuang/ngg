package gterm

import (
	"bytes"
	"crypto/rand"
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

func tryDecode(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	buf, err1 := ss.Base64().DecodeBytes(data)
	val, err2 := hex.DecodeString(string(data))

	if err1 == nil && err2 != nil {
		return buf.Bytes()
	} else if err1 != nil && err2 == nil {
		return val
	} else if err1 != nil && err2 != nil {
		return data
	} else { // err1 == nil && err2 == nil
		chosen, err := gum.Choose([]string{"Hex", "Base64", "Raw"}, gum.ChooseHeader("format "+string(data)))
		if err != nil {
			panic(err)
		}
		switch chosen[0] {
		case "Hex":
			return val
		case "Base64":
			return buf.Bytes()
		default:
			return data
		}
	}
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
			line, err := gum.Input("Input:", gum.InputPlaceholder("Input something"))
			if err != nil {
				return nil, err
			}

			format := SplitSchema(&line, "hex", "base64")

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

	format := SplitSchema(&input, "hex", "base64")

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

func SplitSchema(s *string, allowedSchemas ...string) (tail string) {
	for _, c := range allowedSchemas {
		if strings.HasPrefix(*s, c+"://") {
			*s = (*s)[len(c)+3:]
			return c
		}
	}

	return ""
}

func DecodeBySchema(s string, allowedLen ...int) ([]byte, error) {
	format := SplitSchema(&s, "hex", "base64")
	switch format {
	case "hex":
		return hex.DecodeString(s)
	case "base64":
		result := ss.Base64().Decode(s)
		return result.V1.Bytes(), result.V2
	default:
		return []byte(s), nil
	}
}

func decorateReader(r io.Reader, schema string) io.Reader {
	switch strings.ToLower(schema) {
	case "hex":
		return hex.NewDecoder(r)
	case "base64":
		return ss.NewBase64Reader(r)
	default:
		return r
	}
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
