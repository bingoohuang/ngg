package files

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
)

func GetLine(filename string, lineCh chan string, errCh chan error) {
	file, err := os.Open(filename)
	if err != nil {
		close(lineCh) // close causes range on channel to break out of loop
		errCh <- err
		close(errCh)
		return
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineCh <- scanner.Text()
	}
	close(lineCh) // close causes range on channel to break out of loop
	errCh <- scanner.Err()
	close(errCh)
}

type LineScanner struct {
	scanner *bufio.Scanner
	buf     string
	Stopped bool
}

func NewLineScanner(r io.Reader) *LineScanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(ScanLines)

	return &LineScanner{
		scanner: scanner,
	}
}

func (l *LineScanner) Scan(sep string) (result []string, err error) {
	for l.scanner.Scan() {
		t := l.scanner.Text()
		p := strings.TrimSpace(t)

		// 忽略注释
		if strings.HasPrefix(p, "--") {
			continue
		}
		// 看是否设置命令 %
		if strings.HasPrefix(p, "%") {
			result = l.appendBuf(result)
			result = append(result, p)
			return result, nil
		}

		// 看是否 \G \j \J \I \P 结尾
		if len(p) >= 2 && t[len(p)-2] == '\\' && anyOf(p[len(p)-1], 'G', 'j', 'J', 'I', 'P') {
			tmp := strings.TrimSpace(l.buf + t)
			l.buf = ""
			return []string{tmp}, nil
		}

		if strings.HasPrefix(p, sep) {
			if l.buf != "" {
				result = append(result, l.buf)
				l.buf = ""
			}
			p = strings.TrimPrefix(p, sep)
			if strings.TrimSpace(p) == "" {
				continue
			}
		}
		if strings.HasSuffix(p, sep) {
			tmp := strings.TrimSpace(l.buf + t)
			tmp = strings.TrimSpace(tmp)
			tmp = strings.TrimSuffix(tmp, sep)
			l.buf = ""
			if tmp != "" {
				return []string{tmp}, nil
			}
		}

		l.buf += t
	}

	l.Stopped = true

	if l.buf != "" {
		result = append(result, l.buf)
		l.buf = ""
	}

	return result, l.scanner.Err()
}

func (l *LineScanner) appendBuf(result []string) []string {
	if l.buf != "" {
		b := strings.TrimSpace(l.buf)
		if b != "" {
			result = append(result, b)
		}
		l.buf = ""
	}
	return result
}

func anyOf[T comparable](a T, bb ...T) bool {
	for _, b := range bb {
		if a == b {
			return true
		}
	}

	return false
}

// ScanLines is a split function for a [Scanner] that returns each line of
// text, stripped of any trailing end-of-line marker. The returned line may
// be empty. The end-of-line marker is one optional carriage return followed
// by one mandatory newline. In regular expression notation, it is `\r?\n`.
// The last non-empty line of input will be returned even if it has no
// newline.
func ScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
