package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"time"

	"github.com/musiclib/internal/db"
	"github.com/musiclib/internal/db/migrations"
	"github.com/musiclib/internal/db/repositories"
)

func runImport(sourcePath string) {
	cfg := loadImportConfig()
	setupLogging()

	slog.Info("importing data", "source", sourcePath, "target", cfg.DatabasePath)

	targetDB, err := db.Open(db.Options{
		DatabasePath: cfg.DatabasePath,
		TursoURL:     cfg.TursoURL,
		TursoToken:   cfg.TursoToken,
	})
	if err != nil {
		slog.Error("failed to open target database", "err", err)
		os.Exit(1)
	}
	defer targetDB.Close()

	if err := migrations.Run(targetDB); err != nil {
		slog.Error("failed to run migrations", "err", err)
		os.Exit(1)
	}

	sourceDB, err := sql.Open("sqlite", sourcePath)
	if err != nil {
		slog.Error("failed to open source database", "err", err)
		os.Exit(1)
	}
	defer sourceDB.Close()

	ctx := context.Background()

	// Step 1: Import scrobbles
	slog.Info("importing scrobbles from source")
	rows, err := sourceDB.QueryContext(ctx, `
		SELECT artist, artist_mbid, album, album_mbid, track, track_mbid, timestamp, loved, url
		FROM scrobbles
		ORDER BY timestamp
	`)
	if err != nil {
		slog.Error("failed to query source scrobbles", "err", err)
		os.Exit(1)
	}
	defer rows.Close()

	sr := repositories.NewScrobbleRepository(targetDB)
	ar := repositories.NewArtistRepository(targetDB)
	alr := repositories.NewAlbumRepository(targetDB)
	trr := repositories.NewTrackRepository(targetDB)

	batch := make([]repositories.ScrobbleRow, 0, 500)
	total := 0
	inserted := 0

	for rows.Next() {
		var s repositories.ScrobbleRow
		var loved sql.NullInt64
		var url sql.NullString
		var album, albumMBID sql.NullString
		var artistMBID, trackMBID sql.NullString

		if err := rows.Scan(
			&s.Artist, &artistMBID, &album, &albumMBID,
			&s.Track, &trackMBID, &s.Timestamp, &loved, &url,
		); err != nil {
			slog.Error("failed to scan scrobble", "err", err)
			continue
		}

		s.ArtistMBID = artistMBID
		s.Album = album
		s.AlbumMBID = albumMBID
		s.TrackMBID = trackMBID
		s.Loved = loved
		s.URL = url

		batch = append(batch, s)
		total++

		if len(batch) >= 500 {
			n, err := sr.InsertBatch(ctx, batch)
			if err != nil {
				slog.Error("failed to insert batch", "err", err)
			}
			inserted += n
			batch = batch[:0]

			if total%5000 == 0 {
				slog.Info("import progress", "total", total, "inserted", inserted)
			}
		}
	}

	if len(batch) > 0 {
		n, err := sr.InsertBatch(ctx, batch)
		if err != nil {
			slog.Error("failed to insert final batch", "err", err)
		}
		inserted += n
	}

	slog.Info("scrobbles imported", "total_scanned", total, "inserted", inserted)

	// Step 2: Build artist/album/track entities from scrobbles
	slog.Info("building entities from scrobbles")
	buildEntities(ctx, targetDB, ar, alr, trr)

	// Step 3: Link scrobbles to tracks
	slog.Info("linking scrobbles to tracks")
	linked, err := sr.LinkToTracks(ctx)
	if err != nil {
		slog.Error("failed to link scrobbles", "err", err)
	} else {
		slog.Info("scrobbles linked to tracks", "linked", linked)
	}

	// Step 4: Rebuild FTS
	slog.Info("rebuilding FTS indexes")
	if _, err := targetDB.ExecContext(ctx, `INSERT INTO artists_fts(artists_fts) VALUES('rebuild')`); err != nil {
		slog.Warn("failed to rebuild artists fts", "err", err)
	}
	if _, err := targetDB.ExecContext(ctx, `INSERT INTO albums_fts(albums_fts) VALUES('rebuild')`); err != nil {
		slog.Warn("failed to rebuild albums fts", "err", err)
	}
	if _, err := targetDB.ExecContext(ctx, `INSERT INTO tracks_fts(tracks_fts) VALUES('rebuild')`); err != nil {
		slog.Warn("failed to rebuild tracks fts", "err", err)
	}

	slog.Info("import complete")
}

func buildEntities(ctx context.Context, targetDB *sql.DB, ar *repositories.ArtistRepository, alr *repositories.AlbumRepository, trr *repositories.TrackRepository) {
	now := time.Now().Unix()

	// Collect distinct artists from scrobbles
	type artistInfo struct {
		Name string
		MBID sql.NullString
	}
	var artists []artistInfo
	rows, err := targetDB.QueryContext(ctx, `SELECT DISTINCT artist, artist_mbid FROM scrobbles`)
	if err != nil {
		slog.Error("failed to query distinct artists", "err", err)
		return
	}
	for rows.Next() {
		var a artistInfo
		if err := rows.Scan(&a.Name, &a.MBID); err == nil {
			artists = append(artists, a)
		}
	}
	rows.Close()

	// Batch insert artists
	tx, err := targetDB.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("begin tx", "err", err)
		return
	}
	stmt, _ := tx.PrepareContext(ctx, `INSERT OR IGNORE INTO artists (name, mbid, note, favorite, created_at, updated_at) VALUES (?, ?, '', 0, ?, ?)`)
	for _, a := range artists {
		stmt.ExecContext(ctx, a.Name, a.MBID, now, now)
	}
	stmt.Close()
	tx.Commit()

	// Build artist name -> ID map
	artistMap := make(map[string]int64, len(artists))
	artistRows, _ := targetDB.QueryContext(ctx, `SELECT id, name FROM artists`)
	for artistRows.Next() {
		var id int64
		var name string
		if err := artistRows.Scan(&id, &name); err == nil {
			artistMap[name] = id
		}
	}
	artistRows.Close()
	slog.Info("artists ready", "count", len(artistMap))

	// Collect distinct albums
	type albumInfo struct {
		ArtistName string
		AlbumName  string
		MBID       sql.NullString
	}
	var albums []albumInfo
	albumRows, err := targetDB.QueryContext(ctx, `SELECT DISTINCT artist, album, album_mbid FROM scrobbles WHERE album IS NOT NULL AND album != ''`)
	if err != nil {
		slog.Error("failed to query distinct albums", "err", err)
		return
	}
	for albumRows.Next() {
		var a albumInfo
		if err := albumRows.Scan(&a.ArtistName, &a.AlbumName, &a.MBID); err == nil {
			albums = append(albums, a)
		}
	}
	albumRows.Close()

	// Batch insert albums
	tx, _ = targetDB.BeginTx(ctx, nil)
	stmt, _ = tx.PrepareContext(ctx, `INSERT OR IGNORE INTO albums (artist_id, name, mbid, note, favorite, created_at, updated_at) VALUES (?, ?, ?, '', 0, ?, ?)`)
	for _, al := range albums {
		artistID, ok := artistMap[al.ArtistName]
		if !ok {
			continue
		}
		stmt.ExecContext(ctx, artistID, al.AlbumName, al.MBID, now, now)
	}
	stmt.Close()
	tx.Commit()

	// Build "artistName|albumName" -> album ID map
	albumMap := make(map[string]int64)
	albRows, _ := targetDB.QueryContext(ctx, `
		SELECT al.id, a.name, al.name 
		FROM albums al 
		JOIN artists a ON a.id = al.artist_id
	`)
	for albRows.Next() {
		var id int64
		var artistName, albumName string
		if err := albRows.Scan(&id, &artistName, &albumName); err == nil {
			albumMap[artistName+"|"+albumName] = id
		}
	}
	albRows.Close()
	slog.Info("albums ready", "count", len(albumMap))

	// Collect distinct tracks
	type trackInfo struct {
		ArtistName string
		AlbumName  sql.NullString
		TrackName  string
		MBID       sql.NullString
	}
	var tracks []trackInfo
	trackRows, err := targetDB.QueryContext(ctx, `SELECT DISTINCT artist, album, track, track_mbid FROM scrobbles`)
	if err != nil {
		slog.Error("failed to query distinct tracks", "err", err)
		return
	}
	for trackRows.Next() {
		var t trackInfo
		if err := trackRows.Scan(&t.ArtistName, &t.AlbumName, &t.TrackName, &t.MBID); err == nil {
			tracks = append(tracks, t)
		}
	}
	trackRows.Close()

	// Batch insert tracks
	tx, _ = targetDB.BeginTx(ctx, nil)
	stmt, _ = tx.PrepareContext(ctx, `INSERT OR IGNORE INTO tracks (artist_id, album_id, name, mbid, note, favorite, created_at, updated_at) VALUES (?, ?, ?, ?, '', 0, ?, ?)`)
	for _, t := range tracks {
		artistID, ok := artistMap[t.ArtistName]
		if !ok {
			continue
		}
		var albumID sql.NullInt64
		if t.AlbumName.Valid && t.AlbumName.String != "" {
			if aid, ok := albumMap[t.ArtistName+"|"+t.AlbumName.String]; ok {
				albumID = sql.NullInt64{Int64: aid, Valid: true}
			}
		}
		stmt.ExecContext(ctx, artistID, albumID, t.TrackName, t.MBID, now, now)
	}
	stmt.Close()
	tx.Commit()

	slog.Info("entities built", "artists", len(artistMap), "albums", len(albumMap), "tracks", len(tracks))
}

func loadImportConfig() *db.Options {
	return &db.Options{
		DatabasePath: envOrDefault("DATABASE_PATH", "musiclib.db"),
		TursoURL:     os.Getenv("TURSO_DATABASE_URL"),
		TursoToken:   os.Getenv("TURSO_AUTH_TOKEN"),
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}


