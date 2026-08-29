package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type TrackRow struct {
	ID        int64
	ArtistID  int64
	AlbumID   sql.NullInt64
	Name      string
	MBID      sql.NullString
	Note      string
	Favorite  bool
	CreatedAt int64
	UpdatedAt int64
	ArtistName sql.NullString
	AlbumName  sql.NullString
	ScrobbleCount int
	FirstListened sql.NullInt64
	LastListened  sql.NullInt64
}

type TrackRepository struct {
	db *sql.DB
}

func NewTrackRepository(db *sql.DB) *TrackRepository {
	return &TrackRepository{db: db}
}

func trackSortClause(sort string) string {
	switch sort {
	case "name_desc":
		return "t.name DESC"
	case "scrobbles_desc":
		return "COALESCE(s.scrobble_count, 0) DESC, t.name ASC"
	case "scrobbles_asc":
		return "COALESCE(s.scrobble_count, 0) ASC, t.name ASC"
	case "recent_desc":
		return "s.last_listened DESC NULLS LAST, t.name ASC"
	case "recent_asc":
		return "s.last_listened ASC NULLS FIRST, t.name ASC"
	default:
		return "t.name ASC"
	}
}

const trackListQuery = `
	SELECT
		t.id, t.artist_id, t.album_id, t.name, t.mbid,
		t.note, t.favorite, t.created_at, t.updated_at,
		a.name,
		al.name,
		COALESCE(s.scrobble_count, 0),
		s.first_listened,
		s.last_listened
	FROM tracks t
	JOIN artists a ON a.id = t.artist_id
	LEFT JOIN albums al ON al.id = t.album_id
	LEFT JOIN (
		SELECT sc.track, sc.artist, COUNT(*) as scrobble_count,
			MIN(sc.timestamp) as first_listened,
			MAX(sc.timestamp) as last_listened
		FROM scrobbles sc
		GROUP BY sc.artist, sc.track
	) s ON s.artist = a.name AND s.track = t.name
`

func (r *TrackRepository) List(ctx context.Context, page, limit int, sort string) ([]TrackRow, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tracks`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count tracks: %w", err)
	}

	query := trackListQuery + ` ORDER BY ` + trackSortClause(sort) + ` LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list tracks: %w", err)
	}
	defer rows.Close()

	var tracks []TrackRow
	for rows.Next() {
		var tr TrackRow
		if err := rows.Scan(
			&tr.ID, &tr.ArtistID, &tr.AlbumID, &tr.Name, &tr.MBID,
			&tr.Note, &tr.Favorite, &tr.CreatedAt, &tr.UpdatedAt,
			&tr.ArtistName, &tr.AlbumName,
			&tr.ScrobbleCount, &tr.FirstListened, &tr.LastListened,
		); err != nil {
			return nil, 0, fmt.Errorf("scan track: %w", err)
		}
		tracks = append(tracks, tr)
	}

	return tracks, total, rows.Err()
}

func (r *TrackRepository) GetByID(ctx context.Context, id int64) (*TrackRow, error) {
	var tr TrackRow
	err := r.db.QueryRowContext(ctx, trackListQuery+` WHERE t.id = ?`, id).Scan(
		&tr.ID, &tr.ArtistID, &tr.AlbumID, &tr.Name, &tr.MBID,
		&tr.Note, &tr.Favorite, &tr.CreatedAt, &tr.UpdatedAt,
		&tr.ArtistName, &tr.AlbumName,
		&tr.ScrobbleCount, &tr.FirstListened, &tr.LastListened,
	)
	if err != nil {
		return nil, fmt.Errorf("get track: %w", err)
	}
	return &tr, nil
}

func (r *TrackRepository) FindOrCreate(ctx context.Context, artistID int64, albumID *int64, name string, mbid *string, now int64) (int64, error) {
	var id int64
	var albumIDVal sql.NullInt64
	if albumID != nil {
		albumIDVal = sql.NullInt64{Int64: *albumID, Valid: true}
	}

	err := r.db.QueryRowContext(ctx, `SELECT id FROM tracks WHERE artist_id = ? AND name = ? AND (album_id = ? OR (album_id IS NULL AND ? IS NULL))`,
		artistID, name, albumIDVal, albumIDVal).Scan(&id)
	if err == nil {
		if mbid != nil {
			_, _ = r.db.ExecContext(ctx, `UPDATE tracks SET mbid = ? WHERE id = ? AND mbid IS NULL`, *mbid, id)
		}
		return id, nil
	}

	var mbidVal sql.NullString
	if mbid != nil {
		mbidVal = sql.NullString{String: *mbid, Valid: true}
	}

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO tracks (artist_id, album_id, name, mbid, note, favorite, created_at, updated_at)
		VALUES (?, ?, ?, ?, '', 0, ?, ?)
	`, artistID, albumIDVal, name, mbidVal, now, now)
	if err != nil {
		return 0, fmt.Errorf("insert track: %w", err)
	}

	return result.LastInsertId()
}

func (r *TrackRepository) Update(ctx context.Context, id int64, note *string, favorite *bool) error {
	if note != nil {
		if _, err := r.db.ExecContext(ctx, `UPDATE tracks SET note = ?, updated_at = ? WHERE id = ?`, *note, time.Now().Unix(), id); err != nil {
			return fmt.Errorf("update track note: %w", err)
		}
	}
	if favorite != nil {
		if _, err := r.db.ExecContext(ctx, `UPDATE tracks SET favorite = ?, updated_at = ? WHERE id = ?`, *favorite, time.Now().Unix(), id); err != nil {
			return fmt.Errorf("update track favorite: %w", err)
		}
	}
	return nil
}

func (r *TrackRepository) Search(ctx context.Context, query string, limit int) ([]TrackRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.artist_id, t.album_id, t.name, t.mbid,
			t.note, t.favorite, t.created_at, t.updated_at,
			a.name,
			al.name,
			COALESCE(s.scrobble_count, 0),
			s.first_listened,
			s.last_listened
		FROM tracks t
		JOIN artists a ON a.id = t.artist_id
		LEFT JOIN albums al ON al.id = t.album_id
		JOIN tracks_fts fts ON fts.rowid = t.id
		LEFT JOIN (
			SELECT sc.track, sc.artist, COUNT(*) as scrobble_count,
				MIN(sc.timestamp) as first_listened,
				MAX(sc.timestamp) as last_listened
			FROM scrobbles sc
			GROUP BY sc.artist, sc.track
		) s ON s.artist = a.name AND s.track = t.name
		WHERE tracks_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search tracks: %w", err)
	}
	defer rows.Close()

	var tracks []TrackRow
	for rows.Next() {
		var tr TrackRow
		if err := rows.Scan(&tr.ID, &tr.ArtistID, &tr.AlbumID, &tr.Name, &tr.MBID,
			&tr.Note, &tr.Favorite, &tr.CreatedAt, &tr.UpdatedAt,
			&tr.ArtistName, &tr.AlbumName,
			&tr.ScrobbleCount, &tr.FirstListened, &tr.LastListened); err != nil {
			return nil, fmt.Errorf("scan track: %w", err)
		}
		tracks = append(tracks, tr)
	}
	return tracks, rows.Err()
}

func (r *TrackRepository) GetHistory(ctx context.Context, trackID int64) ([]ScrobbleRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, track_id, artist, artist_mbid, album, album_mbid,
			track, track_mbid, timestamp, loved, url
		FROM scrobbles
		WHERE track_id = ?
		ORDER BY timestamp DESC
	`, trackID)
	if err != nil {
		return nil, fmt.Errorf("get track history: %w", err)
	}
	defer rows.Close()

	var scrobbles []ScrobbleRow
	for rows.Next() {
		var s ScrobbleRow
		if err := rows.Scan(
			&s.ID, &s.TrackID, &s.Artist, &s.ArtistMBID, &s.Album, &s.AlbumMBID,
			&s.Track, &s.TrackMBID, &s.Timestamp, &s.Loved, &s.URL,
		); err != nil {
			return nil, fmt.Errorf("scan scrobble: %w", err)
		}
		scrobbles = append(scrobbles, s)
	}
	return scrobbles, rows.Err()
}

func (r *TrackRepository) SyncFTS(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tracks_fts(tracks_fts) VALUES('rebuild')
	`)
	return err
}
