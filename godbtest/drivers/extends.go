package drivers

import (
	"database/sql"
	"net/url"
)

type DBFixer func(db *sql.DB) error

type SchemeUpdateFn func(u *url.URL) DBFixer

var schemeUpdaters = make(map[string]SchemeUpdateFn)

func RegisterSchemeUpdate(driver string, f SchemeUpdateFn) {
	schemeUpdaters[driver] = f
}

func FixURL(u *url.URL) DBFixer {
	if f, ok := schemeUpdaters[u.Scheme]; ok {
		return f(u)
	}
	return nil
}
