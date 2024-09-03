package files

import (
	"regexp"

	"github.com/emirpasic/gods/maps/linkedhashmap"
)

var (
	key1Pattern = regexp.MustCompile(`([._\-\w]+): `)
	key2Pattern = regexp.MustCompile(`,? ([._\-\w]+): `)
)

// ParseLogLine parse log line like this:
// I0525 15:44:00 20851 see_dat.go:36] fileID: 13,a6cdba7a7cf7, offset: 25798875592, size: 1401024(1.34 MiB) cookie: ba7a7cf7 appendedAt: 2023-05-25 14:53:12.095829 +0800 CST
// to  {"fileID":"13,a6cdba7a7cf7","offset":"25798875592","size":"1401024(1.34 MiB)","cookie":"ba7a7cf7","appendedAt":"2023-05-25 14:53:12.095829 +0800 CST"}
func ParseLogLine(logline string) *linkedhashmap.Map {
	m := linkedhashmap.New()
	subs := key1Pattern.FindStringSubmatchIndex(logline)
	if len(subs) == 0 {
		return m
	}

	key := logline[subs[2]:subs[3]]
	for {
		logline = logline[subs[1]:]
		subs = key2Pattern.FindStringSubmatchIndex(logline)
		if len(subs) == 0 {
			m.Put(key, logline)
			break
		}

		m.Put(key, logline[:subs[0]])
		key = logline[subs[2]:subs[3]]
	}

	return m
}
