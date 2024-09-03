//go:build (all || sqlite3) && !no_sqlite3

package drivers

import (
	_ "github.com/mattn/go-sqlite3"
)
