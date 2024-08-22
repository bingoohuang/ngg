package ss

import (
	"bytes"
	"encoding/base64"
	"io"
	"strings"
)

type base64Reader struct{ io.Reader }

// StdEncoding：RFC 4648 定义的标准 BASE64 编码字符集，结果填充=，使字节数为4的倍数
// URLEncoding：RFC 4648 定义的另一 BASE64 编码字符集，用 - 和 _ 替换了 + 和 /，用于URL和文件名，结果填充=
// RawStdEncoding：同 StdEncoding，但结果不填充=
// RawURLEncoding：同 URLEncoding，但结果不填充=
func (f *base64Reader) Read(p []byte) (int, error) {
	n, err := f.Reader.Read(p)

	for i := 0; i < n; i++ {
		switch p[i] {
		case '-':
			p[i] = '+'
		case '_':
			p[i] = '/'
		case '=':
			n = i
			return n, err
		}
	}

	return n, err
}

// Decode copies io.Reader which is in base64 format ( any one of StdEncoding/URLEncoding/RawStdEncoding/RawURLEncoding).
func DecodeBase64(dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(dst, base64.NewDecoder(base64.RawStdEncoding, &base64Reader{Reader: src}))
}

func Base64() *b64 {
	return &b64{}
}

// EncodeBytes encodes src into base64 []byte.
func (b *b64) EncodeBytes(src []byte, flags ...Base64Flags) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if _, err := EncodeBase64(&buf, bytes.NewReader(src), flags...); err != nil {
		return nil, err
	}
	return &buf, nil
}

// Encode encodes src into base64 string.
func (b *b64) Encode(src string, flags ...Base64Flags) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if _, err := EncodeBase64(&buf, strings.NewReader(src), flags...); err != nil {
		return nil, err
	}
	return &buf, nil
}

func (b *b64) Decode(src string) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if _, err := DecodeBase64(&buf, strings.NewReader(src)); err != nil {
		return nil, err
	}
	return &buf, nil
}

// DecodeBytes decode bytes which is in base64 format ( any one of StdEncoding/URLEncoding/RawStdEncoding/RawURLEncoding).
func (b *b64) DecodeBytes(src []byte) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if _, err := DecodeBase64(&buf, bytes.NewReader(src)); err != nil {
		return nil, err
	}
	return &buf, nil
}

type b64 struct{}

type Base64Flags uint8

const (
	Std Base64Flags = 1 << iota
	Url
	Raw
)

// EncodeBase64 copies io.Reader to io.Writer which is in base64 format ( any one of StdEncoding/URLEncoding/RawStdEncoding/RawURLEncoding).
func EncodeBase64(dst io.Writer, src io.Reader, flags ...Base64Flags) (int64, error) {
	enc := base64.StdEncoding
	var flag Base64Flags
	for _, f := range flags {
		flag |= f
	}

	switch {
	case flag&Url == Url:
		enc = If(flag&Raw == Raw, base64.RawURLEncoding, base64.URLEncoding)
	case flag&Raw == Raw:
		enc = base64.RawStdEncoding
	}

	closer := base64.NewEncoder(enc, dst)
	defer closer.Close()
	return io.Copy(closer, src)
}
