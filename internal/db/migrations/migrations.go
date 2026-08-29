package migrations

import (
	"database/sql"
	"fmt"
	"log/slog"
	"sort"
	"time"
)

type Migration struct {
	Version int
	Name    string
	SQL     string
}

func All() []Migration {
	return []Migration{
		{
			Version: 1,
			Name:    "create_artists",
			SQL: `
				CREATE TABLE IF NOT EXISTS artists (
					id INTEGER PRIMARY KEY,
					name TEXT NOT NULL,
					mbid TEXT UNIQUE,
					note TEXT NOT NULL DEFAULT '',
					favorite INTEGER NOT NULL DEFAULT 0,
					created_at INTEGER NOT NULL,
					updated_at INTEGER NOT NULL
				);
				CREATE INDEX IF NOT EXISTS idx_artists_name ON artists(name);
			`,
		},
		{
			Version: 2,
			Name:    "create_albums",
			SQL: `
				CREATE TABLE IF NOT EXISTS albums (
					id INTEGER PRIMARY KEY,
					artist_id INTEGER NOT NULL,
					name TEXT NOT NULL,
					mbid TEXT,
					release_date TEXT,
					note TEXT NOT NULL DEFAULT '',
					favorite INTEGER NOT NULL DEFAULT 0,
					created_at INTEGER NOT NULL,
					updated_at INTEGER NOT NULL,
					FOREIGN KEY (artist_id) REFERENCES artists(id) ON DELETE CASCADE,
					UNIQUE(artist_id, name)
				);
				CREATE INDEX IF NOT EXISTS idx_albums_artist ON albums(artist_id);
			`,
		},
		{
			Version: 3,
			Name:    "create_tracks",
			SQL: `
				CREATE TABLE IF NOT EXISTS tracks (
					id INTEGER PRIMARY KEY,
					artist_id INTEGER NOT NULL,
					album_id INTEGER,
					name TEXT NOT NULL,
					mbid TEXT,
					note TEXT NOT NULL DEFAULT '',
					favorite INTEGER NOT NULL DEFAULT 0,
					created_at INTEGER NOT NULL,
					updated_at INTEGER NOT NULL,
					FOREIGN KEY (artist_id) REFERENCES artists(id) ON DELETE CASCADE,
					FOREIGN KEY (album_id) REFERENCES albums(id) ON DELETE SET NULL,
					UNIQUE(artist_id, album_id, name)
				);
				CREATE INDEX IF NOT EXISTS idx_tracks_artist ON tracks(artist_id);
				CREATE INDEX IF NOT EXISTS idx_tracks_album ON tracks(album_id);
			`,
		},
		{
			Version: 4,
			Name:    "create_scrobbles",
			SQL: `
				CREATE TABLE IF NOT EXISTS scrobbles (
					id INTEGER PRIMARY KEY,
					track_id INTEGER,
					artist TEXT NOT NULL,
					artist_mbid TEXT,
					album TEXT,
					album_mbid TEXT,
					track TEXT NOT NULL,
					track_mbid TEXT,
					timestamp INTEGER NOT NULL,
					loved INTEGER,
					url TEXT,
					FOREIGN KEY (track_id) REFERENCES tracks(id) ON DELETE SET NULL,
					UNIQUE(artist, track, timestamp)
				);
				CREATE INDEX IF NOT EXISTS idx_scrobbles_track ON scrobbles(track_id);
				CREATE INDEX IF NOT EXISTS idx_scrobbles_timestamp ON scrobbles(timestamp);
			`,
		},
		{
			Version: 5,
			Name:    "create_tags",
			SQL: `
				CREATE TABLE IF NOT EXISTS tags (
					id INTEGER PRIMARY KEY,
					name TEXT NOT NULL UNIQUE
				);
				CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);

				CREATE TABLE IF NOT EXISTS entity_tags (
					id INTEGER PRIMARY KEY,
					tag_id INTEGER NOT NULL,
					entity_type TEXT NOT NULL,
					entity_id INTEGER NOT NULL,
					created_at INTEGER NOT NULL,
					FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,
					UNIQUE(tag_id, entity_type, entity_id)
				);
				CREATE INDEX IF NOT EXISTS idx_entity_tags_entity ON entity_tags(entity_type, entity_id);
			`,
		},
		{
			Version: 6,
			Name:    "create_fts",
			SQL: `
				CREATE VIRTUAL TABLE IF NOT EXISTS artists_fts USING fts5(
					name,
					content=artists,
					content_rowid=id
				);
				CREATE VIRTUAL TABLE IF NOT EXISTS albums_fts USING fts5(
					name,
					content=albums,
					content_rowid=id
				);
				CREATE VIRTUAL TABLE IF NOT EXISTS tracks_fts USING fts5(
					name,
					content=tracks,
					content_rowid=id
				);
			`,
		},
		{
			Version: 7,
			Name:    "create_schema_migrations",
			SQL: `
				CREATE TABLE IF NOT EXISTS schema_migrations (
					version INTEGER PRIMARY KEY,
					name TEXT NOT NULL,
					applied_at INTEGER NOT NULL
				);
			`,
		},
		{
			Version: 8,
			Name:    "create_lastfm_state",
			SQL: `
				CREATE TABLE IF NOT EXISTS lastfm_sync_state (
					id INTEGER PRIMARY KEY CHECK (id = 1),
					last_sync_at INTEGER,
					last_page INTEGER DEFAULT 0,
					total_imported INTEGER DEFAULT 0
				);
			`,
		},
		{
			Version: 9,
			Name:    "create_collections",
			SQL: `
				CREATE TABLE IF NOT EXISTS collections (
					id INTEGER PRIMARY KEY,
					parent_id INTEGER,
					name TEXT NOT NULL,
					description TEXT NOT NULL DEFAULT '',
					created_at INTEGER NOT NULL,
					updated_at INTEGER NOT NULL,
					FOREIGN KEY (parent_id) REFERENCES collections(id) ON DELETE CASCADE
				);
				CREATE INDEX IF NOT EXISTS idx_collections_parent ON collections(parent_id);

				CREATE TABLE IF NOT EXISTS collection_items (
					id INTEGER PRIMARY KEY,
					collection_id INTEGER NOT NULL,
					item_type TEXT NOT NULL,
					item_id INTEGER NOT NULL,
					position INTEGER NOT NULL DEFAULT 0,
					note TEXT NOT NULL DEFAULT '',
					created_at INTEGER NOT NULL,
					FOREIGN KEY (collection_id) REFERENCES collections(id) ON DELETE CASCADE,
					CHECK (item_type IN ('artist', 'album', 'track', 'collection'))
				);
				CREATE INDEX IF NOT EXISTS idx_collection_items_collection ON collection_items(collection_id);
				CREATE INDEX IF NOT EXISTS idx_collection_items_position ON collection_items(collection_id, position);
				CREATE UNIQUE INDEX IF NOT EXISTS idx_collection_items_unique ON collection_items(collection_id, item_type, item_id);
			`,
		},
	}
}

func Run(db *sql.DB) error {
	slog.Info("running migrations")

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at INTEGER NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("ensure migrations table: %w", err)
	}

	applied, err := getApplied(db)
	if err != nil {
		return fmt.Errorf("get applied migrations: %w", err)
	}

	migrations := All()
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}

		slog.Info("applying migration", "version", m.Version, "name", m.Name)

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", m.Version, err)
		}

		if _, err := tx.Exec(m.SQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %d (%s): %w", m.Version, m.Name, err)
		}

		if _, err := tx.Exec(
			`INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)`,
			m.Version, m.Name, nowUnix(),
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %d: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.Version, err)
		}

		slog.Info("migration applied", "version", m.Version, "name", m.Name)
	}

	slog.Info("migrations complete")
	return nil
}

func getApplied(db *sql.DB) (map[int]bool, error) {
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

func nowUnix() int64 {
	return time.Now().Unix()
}
