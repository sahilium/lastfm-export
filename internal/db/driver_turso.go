//go:build turso

package db

import (
	_ "github.com/tursodatabase/go-libsql"
	_ "modernc.org/sqlite"
)
