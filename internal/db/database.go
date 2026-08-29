package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
)

type Options struct {
	DatabasePath string
	TursoURL     string
	TursoToken   string
}

func Open(opts Options) (*sql.DB, error) {
	var dsn string
	var driver string

	if opts.TursoURL != "" {
		driver = "libsql"
		dsn = opts.TursoURL
		if opts.TursoToken != "" {
			dsn += "?authToken=" + opts.TursoToken
		}
		slog.Info("connecting to turso", "url", redactURL(opts.TursoURL))
	} else {
		driver = "sqlite"
		dsn = fmt.Sprintf(
			"file:%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON&_synchronous=NORMAL",
			opts.DatabasePath,
		)
		slog.Info("opening local sqlite", "path", opts.DatabasePath)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	slog.Info("database connected", "driver", driver)
	return db, nil
}

func redactURL(url string) string {
	if idx := strings.Index(url, "?"); idx != -1 {
		return url[:idx]
	}
	return url
}
