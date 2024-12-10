package randpoem

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"io"
	"log"

	"github.com/bingoohuang/ngg/jj"
	"github.com/bingoohuang/ngg/ss"
)

func AdaptEncoding(val, options string) string {
	arg := struct {
		Base64 bool
		Url    bool
		Raw    bool
		Hex    bool
	}{}

	jj.ParseConf(options, &arg)
	switch {
	case arg.Base64:
		var flags []ss.Base64Flags
		if arg.Url {
			flags = append(flags, ss.Url)
		}
		if arg.Raw {
			flags = append(flags, ss.Raw)
		}

		return ss.Base64().Encode(val, flags...).V1.String()
	case arg.Hex:
		return hex.EncodeToString([]byte(val))
	}

	return val
}

func SliceRandItem(data []string) string {
	return data[ss.Rand().Intn(len(data))]
}

func UnGzipLines(input []byte) []string {
	content := MustUnGzip(input)
	scanner := bufio.NewScanner(bytes.NewReader(content))

	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return lines
}

func MustUnGzip(input []byte) []byte {
	r, err := UnGzip(input)
	if err != nil {
		log.Fatal(err)
	}
	return r
}

func UnGzip(input []byte) ([]byte, error) {
	g, err := gzip.NewReader(bytes.NewReader(input))
	if err != nil {
		return nil, err
	}
	defer g.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, g)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
