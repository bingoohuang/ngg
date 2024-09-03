//go:build (!no_base || sqlite) && !no_sqlite

package drivers

import (
	_ "modernc.org/sqlite"
)
