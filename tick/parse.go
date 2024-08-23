package tick

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// nolint:gomnd,gochecknoglobals
var unitMap = map[string]int64{
	"ns":     int64(time.Nanosecond),
	"纳秒":     int64(time.Nanosecond),
	"us":     int64(time.Microsecond),
	"µs":     int64(time.Microsecond), // U+00B5 = micro symbol
	"μs":     int64(time.Microsecond), // U+03BC = Greek letter mu
	"微妙":     int64(time.Microsecond), // U+03BC = Greek letter mu
	"ms":     int64(time.Millisecond),
	"毫秒":     int64(time.Millisecond),
	"s":      int64(time.Second),
	"秒":      int64(time.Second),
	"m":      int64(time.Minute),
	"分":      int64(time.Minute),
	"h":      int64(time.Hour),
	"时":      int64(time.Hour),
	"d":      int64(time.Hour * 24),      // a day
	"day":    int64(time.Hour * 24),      // a day
	"days":   int64(time.Hour * 24),      // a day
	"天":      int64(time.Hour * 24),      // a day
	"w":      int64(time.Hour * 24 * 7),  // a week
	"周":      int64(time.Hour * 24 * 7),  // a week
	"week":   int64(time.Hour * 24 * 7),  // a week
	"weeks":  int64(time.Hour * 24 * 7),  // a week
	"month":  int64(30 * 24 * time.Hour), // 月
	"月":      int64(30 * 24 * time.Hour), // 月
	"mon":    int64(30 * 24 * time.Hour), // 月
	"months": int64(30 * 24 * time.Hour), // 月
}

type Fraction struct {
	Unit  string
	Value int64
}

// Parse parses a duration string(add d and w to standard time.Parse).
// A duration string is a possibly signed sequence of
// decimal numbers, each with optional fraction and a unit suffix,
// such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d", "w".
// allowUnits 允许的单位, 不传，使用上面的 unitMap 中的默认单位
func Parse(s string, allowUnits ...string) (time.Duration, []Fraction, error) {
	var allowUnitMap = unitMap
	if len(allowUnits) > 0 {
		allowUnitMap = make(map[string]int64)
		for _, allowUnit := range allowUnits {
			unit := strings.ToLower(allowUnit)
			if val, ok := unitMap[unit]; !ok {
				return 0, nil, fmt.Errorf("unknown unit %s", allowUnit)
			} else {
				allowUnitMap[unit] = val
			}
		}
	}

	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
	orig := s

	var d int64

	neg := false

	// Consume [-+]?
	if s != "" {
		c := s[0]
		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}

	// Special case: if all that is left is "0", this is zero.
	if s == "0" {
		return 0, nil, nil
	}

	if s == "" {
		return 0, nil, errors.New("time: invalid duration " + orig)
	}

	var fractions []Fraction

	for s != "" {
		var (
			v, f  int64       // integers before, after decimal point
			scale float64 = 1 // value = v + f/scale
		)

		var err error

		// The next character must be [0-9.]
		if !(s[0] == '.' || '0' <= s[0] && s[0] <= '9') {
			return 0, nil, errors.New("time: invalid duration " + orig)
		}
		// Consume [0-9]*
		pl := len(s)
		v, s, err = leadingInt(s)

		if err != nil {
			return 0, nil, errors.New("time: invalid duration " + orig)
		}

		pre := pl != len(s) // whether we consumed anything before a period
		// Consume (\.[0-9]*)?
		post := false

		if s != "" && s[0] == '.' {
			s = s[1:]
			pl := len(s)
			f, scale, s = leadingFraction(s)
			post = pl != len(s)
		}

		if !pre && !post {
			// no digits (e.g. ".s" or "-.s")
			return 0, nil, errors.New("time: invalid duration " + orig)
		}

		// Consume unit.
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			if c == '.' || '0' <= c && c <= '9' {
				break
			}
		}

		if i == 0 {
			return 0, nil, errors.New("time: missing unit in duration " + orig)
		}

		u := s[:i]
		s = s[i:]
		unit, ok := allowUnitMap[strings.ToLower(u)]

		if !ok {
			return 0, nil, errors.New("time: unknown unit " + u + " in duration " + orig)
		}

		if v > (1<<63-1)/unit {
			// overflow
			return 0, nil, errors.New("time: invalid duration " + orig)
		}

		fractions = append(fractions, Fraction{Unit: u, Value: v})

		v *= unit

		if f > 0 {
			// float64 is needed to be nanosecond accurate for fractions of hours.
			// v >= 0 && (f*unit/scale) <= 3.6e+12 (ns/h, h is the largest unit)
			v += int64(float64(f) * (float64(unit) / scale))
			if v < 0 {
				// overflow
				return 0, nil, errors.New("time: invalid duration " + orig)
			}
		}

		d += v

		if d < 0 {
			// overflow
			return 0, nil, errors.New("time: invalid duration " + orig)
		}
	}

	if neg {
		d = -d
	}

	return time.Duration(d), fractions, nil
}

var errLeadingInt = errors.New("time: bad [0-9]*") // never printed

// leadingInt consumes the leading [0-9]* from s.
func leadingInt(s string) (x int64, rem string, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}

		if x > (1<<63-1)/10 {
			// overflow
			return 0, "", errLeadingInt
		}

		x = x*10 + int64(c) - '0'
		if x < 0 {
			// overflow
			return 0, "", errLeadingInt
		}
	}

	return x, s[i:], nil
}

// leadingFraction consumes the leading [0-9]* from s.
// It is used only for fractions, so does not return an error on overflow,
// it just stops accumulating precision.
func leadingFraction(s string) (x int64, scale float64, rem string) {
	i := 0
	scale = 1
	overflow := false

	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}

		if overflow {
			continue
		}

		if x > (1<<63-1)/10 {
			// It's possible for overflow to give a positive number, so take care.
			overflow = true
			continue
		}

		y := x*10 + int64(c) - '0'
		if y < 0 {
			overflow = true
			continue
		}

		x = y
		scale *= 10
	}

	return x, scale, s[i:]
}
