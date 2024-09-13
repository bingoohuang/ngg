package encoder

import (
	"context"
	"io"

	"github.com/bingoohuang/ngg/go-json/api"
)

type OptionFlag uint8

const (
	HTMLEscapeOption OptionFlag = 1 << iota
	IndentOption
	UnorderedMapOption
	DebugOption
	ColorizeOption
	ContextOption
	NormalizeUTF8Option
	FieldQueryOption
)

type Option struct {
	Flag           OptionFlag
	ColorScheme    *ColorScheme
	Context        context.Context
	DebugOut       io.Writer
	DebugDOTOut    io.WriteCloser
	NamingStrategy api.NamingStrategy
}

func (o *Option) Reset() {
	o.Flag = 0
	o.ColorScheme = nil
	o.Context = nil
	o.DebugOut = nil
	o.DebugDOTOut = nil
	o.NamingStrategy = func(key string) string { return key }
}

type EncodeFormat struct {
	Header string
	Footer string
}

type EncodeFormatScheme struct {
	Int       EncodeFormat
	Uint      EncodeFormat
	Float     EncodeFormat
	Bool      EncodeFormat
	String    EncodeFormat
	Binary    EncodeFormat
	ObjectKey EncodeFormat
	Null      EncodeFormat
}

type (
	ColorScheme = EncodeFormatScheme
	ColorFormat = EncodeFormat
)
