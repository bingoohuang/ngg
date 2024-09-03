package conf

import (
	"sync"
	"time"
	"unicode"

	"github.com/bingoohuang/ngg/ss"
)

func AbbreviateSlice(r []any, maxLen int) []any {
	if maxLen == 0 {
		return r
	}

	result := make([]any, len(r))
	for i, item := range r {
		result[i] = ss.AbbreviateAny(item, maxLen, "…")
	}
	return result
}

func GoWait(threads int, f func(threadNum int) error, errHandler func(error)) time.Duration {
	var wg sync.WaitGroup
	start := time.Now()
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := f(i); err != nil {
				errHandler(err)
			}
		}(i)
	}

	wg.Wait()
	return time.Since(start)
}

type FindOption struct {
	Open, Close rune
	Quote       rune
}

type Pos struct {
	From, To int
	ArgOpen  int
}

type ParseResult struct {
	Tags  []*Pos
	Names []*Pos
}

func (f *FindOption) FindTags(s string) (tags []*Pos) {
	var pos *Pos

	inQuote := false

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch r {
		case f.Quote:
			inQuote = !inQuote
		case f.Open:
			if !inQuote {
				pos = &Pos{From: i, To: -1}
			}
		case f.Close:
			if !inQuote && pos != nil {
				pos.To = i + 1
				tags = append(tags, pos)
				pos = nil
			}
		}
	}

	return tags
}

func (f *FindOption) FindNamed(s string) (names []*Pos) {
	inQuote := false
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		switch r {
		case f.Quote:
			inQuote = !inQuote
		case ':':
			if !inQuote {
				j := i + 1
				pos := &Pos{From: i, To: -1}
				for ; j < len(runes); j++ {
					next := runes[j]
					if unicode.IsLetter(next) || unicode.IsDigit(next) {
						continue
					}
					break
				}
				if j < len(runes) && runes[j] == '{' {
					k := j + 1
					for ; k < len(runes); k++ {
						if runes[k] == '}' {
							break
						}
					}

					if k < len(runes) {
						pos.To = k + 1
						pos.ArgOpen = j
						names = append(names, pos)

						i = k
						continue
					}
				}

				pos.To = j
				names = append(names, pos)

				i = j
				continue
			}
		}
	}

	return names
}
