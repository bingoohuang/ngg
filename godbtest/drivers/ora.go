//go:build (all || ora) && !no_ora

package drivers

import (
	"database/sql"
	"net/url"
	"strings"

	go_ora "github.com/sijms/go-ora/v2"
	"go.uber.org/multierr"
)

func init() {

	// process url query parameters, e.g. SESSION_CURRENT_SCHEMA=MSP_AUTH
	RegisterSchemeUpdate("oracle", func(u *url.URL) DBFixer {
		sessionParams := make(map[string]string)
		query := u.Query()
		for k, v := range query {
			if kk := findSessionParam(k); kk != "" {
				sessionParams[kk] = v[0]
				delete(query, k)
			}
		}

		if len(sessionParams) == 0 {
			return nil
		}

		u.RawQuery = query.Encode()
		return func(db *sql.DB) error {
			var err error
			for k, v := range sessionParams {
				if e := go_ora.AddSessionParam(db, k, v); e != nil {
					err = multierr.Append(err, e)
				}
			}

			return err
		}
	})
}

var alias = map[string]string{
	"schema": "CURRENT_SCHEMA",
}

func findSessionParam(k string) string {
	if kk := alias[strings.ToLower(k)]; kk != "" {
		return kk
	} else if strings.HasPrefix(k, "SESSION_") {
		return k[8:]
	}

	return ""
}
