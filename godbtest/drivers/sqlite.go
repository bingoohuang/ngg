//go:build (!no_base || sqlite) && !no_sqlite

package drivers

// TODO: panic: sql: Register called twice for driver sqlite
import (
	_ "modernc.org/sqlite"
)
