package ss

import (
	"strings"
)

// Expandable abstract a thing that can be expanded to parts.
type Expandable interface {
	// MakePart returns i'th item.
	ExpandPart(i int) string
	// Len returns the length of part items.
	Len() int
}

// ExpandPart as a part of something.
type ExpandPart struct {
	p []string
}

// MakePart make a direct part by s.
func MakePart(s []string) ExpandPart { return ExpandPart{p: s} }

// MakeFixedPart make a fixed part by s.
func MakeFixedPart(s string) ExpandPart { return ExpandPart{p: []string{s}} }

// Len returns the length of part items.
func (f ExpandPart) Len() int { return len(f.p) }

// ExpandPart returns i'th item.
func (f ExpandPart) ExpandPart(i int) string {
	l := len(f.p)

	if i >= l {
		return f.p[l-1]
	}

	return f.p[i]
}

// MakeExpandPart makes an expanded part.
func MakeExpandPart(s string) ExpandPart {
	expanded := make([]string, 0)
	fs := Fields(s, -1)

	for _, f := range fs {
		items := ExpandRange(f)
		expanded = append(expanded, items...)
	}

	return ExpandPart{p: expanded}
}

// Expand structured a expandable unit.
type Expand struct {
	raw   string
	parts []Expandable
}

// MaxLen returns the max length among the inner parts.
func (f Expand) MaxLen() int {
	maxLen := 0

	for _, p := range f.parts {
		l := p.Len()
		if l > maxLen {
			maxLen = l
		}
	}

	return maxLen
}

// MakePart makes a part of expand.
func (f Expand) MakePart() ExpandPart {
	return MakePart(f.MakeExpand())
}

// MakeExpand makes a expanded string slice of expand.
func (f Expand) MakeExpand() []string {
	ml := f.MaxLen()
	parts := make([]string, ml)

	for i := 0; i < ml; i++ {
		part := ""

		for _, p := range f.parts {
			part += p.ExpandPart(i)
		}

		parts[i] = part
	}

	return parts
}

// MakeExpand  makes an expand by s.
func MakeExpand(s string) Expand {
	parts := make([]Expandable, 0)

	for {
		l := strings.Index(s, "(")
		if l < 0 {
			parts = append(parts, MakeFixedPart(s))
			break
		}

		r := strings.Index(s[l:], ")")
		if r < 0 {
			parts = append(parts, MakeFixedPart(s))
			break
		}

		if lp := s[0:l]; lp != "" {
			parts = append(parts, MakeFixedPart(lp))
		}

		parts = append(parts, MakeExpandPart(s[l+1:l+r]))

		if l+r+1 == len(s) {
			break
		}

		s = s[l+r+1:]
	}

	return Expand{raw: s, parts: parts}
}
