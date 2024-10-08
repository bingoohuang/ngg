package internal

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bingoohuang/ngg/berf/pkg/util"
	"github.com/bingoohuang/ngg/ss"
)

type TagValue interface {
	Contains(s string) bool
}

type Tag struct {
	Raw    string
	Values []TagValue
}

func (t Tag) String() string { return t.Raw }

func (t *Tag) Contains(s string) bool {
	for _, value := range t.Values {
		if value.Contains(s) {
			return true
		}
	}

	return false
}

type SingleValue struct {
	Value string
}

func (v SingleValue) Contains(s string) bool { return strings.EqualFold(v.Value, s) }

type RangeValue struct {
	From string
	To   string

	IsInt   bool
	FromInt int
	ToInt   int
}

func (r *RangeValue) Contains(s string) bool {
	if ss.IsDigits(s) && r.IsInt {
		return r.containsInt(ss.Must(ss.Parse[int](s)))
	}
	return r.containsString(s)
}

func (r *RangeValue) containsInt(v int) bool {
	return (r.From == "" || r.FromInt <= v) && (r.To == "" || v <= r.ToInt)
}

func (r *RangeValue) containsString(v string) bool {
	return (r.From == "" || r.From <= v) && (r.To == "" || v <= r.To)
}

func ParseTag(s string) *Tag {
	tag := &Tag{Raw: s}
	parts := ss.Split(s, ",")
	for _, part := range parts {
		p := strings.Index(part, "-")
		if p < 0 {
			tag.Values = append(tag.Values, &SingleValue{Value: part})
		} else {
			tag.Values = append(tag.Values, NewRangeValue(part[:p], part[p+1:]))
		}
	}

	return tag
}

func NewRangeValue(a, b string) TagValue {
	r := &RangeValue{
		From:    strings.TrimSpace(a),
		To:      strings.TrimSpace(b),
		IsInt:   false,
		FromInt: 0,
		ToInt:   0,
	}

	r.IsInt = (r.From == "" || ss.IsDigits(r.From)) && (r.To == "" || ss.IsDigits(r.To))
	r.FromInt, _ = ss.Parse[int](r.From)
	r.ToInt, _ = ss.Parse[int](r.To)

	if r.From > r.To {
		r.From, r.To = r.To, r.From
	}
	if r.FromInt > r.ToInt {
		r.FromInt, r.ToInt = r.ToInt, r.FromInt
	}

	return r
}

func ParseProfileArg(profileArg []string, envName string) []*Profile {
	var profiles []*Profile
	hasNew := false
	var tag *Tag
	for _, p := range profileArg {

		if util.SplitTail(&p, ":new") {
			ss.ExitIfErr(os.WriteFile(p, DemoProfile, os.ModePerm))
			fmt.Printf("profile file %s created\n", p)
			hasNew = true
			continue
		}

		if tagPos := strings.LastIndex(p, ":"); tagPos > 0 {
			tag = ParseTag(p[tagPos+1:])
			p = p[:tagPos]
		}

		if ok, _ := ss.Exists(p); !ok {
			ss.Exit("profile "+p+" doesn't exist", 1)
		}

		pp, err := ParseProfileFile(p, envName)
		ss.ExitIfErr(err)

		for _, p1 := range pp {
			if tag == nil || tag.Contains(p1.Tag) {
				profiles = append(profiles, p1)
			}
		}
	}
	if hasNew {
		os.Exit(0)
	}

	if len(profileArg) > 0 && len(profiles) == 0 && tag != nil {
		log.Fatalf("failed to find profile with tag %v", tag)
	}

	return profiles
}
