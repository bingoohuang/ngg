//go:build (all || pgx) && !no_pgx

package drivers

import _ "github.com/jackc/pgx/v5/stdlib"
