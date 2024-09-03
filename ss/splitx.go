package ss

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// OpenClose stands for open-close strings like ()[]{} and etc.
type OpenClose struct {
	Open  string
	Close string
}

// IsSame tells the open and close is same of not.
func (o OpenClose) IsSame() bool {
	return o.Open == o.Close
}

type rememberOpenClose struct {
	OpenClose
	startPos int
}

// SplitX splits s by separate (not in (),[],{})
func SplitX(s string, separate string, ocs ...OpenClose) []string {
	ocs = setDefaultOpenCloses(ocs)

	subs := make([]string, 0)

	remembers := make([]rememberOpenClose, 0)
	pos := 0
	l := len(s)

	for i := 0; i < l; {
		// w 当前字符宽度
		runeValue, w := utf8.DecodeRuneInString(s[i:])
		ch := string(runeValue)

		switch {
		case runeValue == '\\':
			if i+w < l {
				_, nextWidth := utf8.DecodeRuneInString(s[i+w:])
				i += nextWidth
			}
		case len(remembers) > 0:
			last := remembers[len(remembers)-1]
			if yes, oc := isOpen(ch, ocs, last.OpenClose); yes {
				remembers = append(remembers, rememberOpenClose{OpenClose: oc, startPos: i})
			} else if ch == last.Close {
				remembers = remembers[0 : len(remembers)-1]
				if len(remembers) == 0 {
					remembers = make([]rememberOpenClose, 0)
				}
			}
		default:
			if yes, oc := isOpen(ch, ocs, OpenClose{}); yes {
				remembers = append(remembers, rememberOpenClose{OpenClose: oc, startPos: i})
			} else if ch == separate {
				subs = tryAddPart(subs, s[pos:i])
				pos = i + w
			}
		}

		i += w
	}

	if pos < l {
		subs = tryAddPart(subs, s[pos:])
	}

	return subs
}

func setDefaultOpenCloses(ocs []OpenClose) []OpenClose {
	if len(ocs) == 0 {
		return []OpenClose{
			{"(", ")"},
			{"{", "}"},
			{"[", "]"},
			{"'", "'"},
		}
	}

	return ocs
}

func isOpen(s string, ocs []OpenClose, last OpenClose) (yes bool, oc OpenClose) {
	for _, oc := range ocs {
		if s == oc.Open {
			if !oc.IsSame() || s != last.Close {
				return true, oc
			}
		}
	}

	return false, OpenClose{}
}

func tryAddPart(subs []string, sub string) []string {
	s := strings.TrimSpace(sub)
	if s != "" {
		return append(subs, s)
	}

	return subs
}

// SplitReg slices s into substrings separated by the expression and returns a slice of
// the substrings between those expression matches. (especially including the tail separator.)
//
// The slice returned by this method consists of all the substrings of s
// not contained in the slice returned by FindAllString. When called on an expression
// that contains no metacharacters, it is equivalent to strings.SplitN.
//
// Example:
//
//	s := regexp.MustCompile("a*").SplitReg("abaabaccadaaae", 5)
//	// s: ["", "b", "b", "c", "cadaaae"]
//
// The count determines the number of substrings to return:
//
//	n > 0: at most n substrings; the last substring will be the unsplit remainder.
//	n == 0: the result is nil (zero substrings)
//	n < 0: all substrings
func SplitReg(s string, re *regexp.Regexp, n int) []string {
	if n == 0 {
		return nil
	}

	if len(re.String()) > 0 && len(s) == 0 {
		return []string{""}
	}

	matches := re.FindAllStringIndex(s, n)
	result := make([]string, 0, len(matches))

	beg, end := 0, 0
	for _, match := range matches {
		if n > 0 && len(result) >= n-1 {
			break
		}

		end = match[0]
		if match[1] != 0 {
			if sub := strings.TrimSpace(s[beg:end]); sub != "" {
				// including the seps
				result = append(result, sub+s[match[0]:match[1]])
			}
		}
		beg = match[1]
	}

	if end != len(s) {
		if sub := strings.TrimSpace(s[beg:]); sub != "" {
			result = append(result, sub)
		}
	}

	return result
}
