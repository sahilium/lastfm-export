package repositories

import (
	"context"
	"database/sql"
	"fmt"
)

type ScrobbleRow struct {
	ID         int64
	TrackID    sql.NullInt64
	Artist     string
	ArtistMBID sql.NullString
	Album      sql.NullString
	AlbumMBID  sql.NullString
	Track      string
	TrackMBID  sql.NullString
	Timestamp  int64
	Loved      sql.NullInt64
	URL        sql.NullString
}

type ScrobbleRepository struct {
	db *sql.DB
}

func NewScrobbleRepository(db *sql.DB) *ScrobbleRepository {
	return &ScrobbleRepository{db: db}
}

func (r *ScrobbleRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM scrobbles`).Scan(&count)
	return count, err
}

func (r *ScrobbleRepository) InsertBatch(ctx context.Context, scrobbles []ScrobbleRow) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR IGNORE INTO scrobbles
			(artist, artist_mbid, album, album_mbid, track, track_mbid, timestamp, loved, url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	inserted := 0
	for _, s := range scrobbles {
		result, err := stmt.ExecContext(ctx,
			s.Artist, s.ArtistMBID, s.Album, s.AlbumMBID,
			s.Track, s.TrackMBID, s.Timestamp, s.Loved, s.URL,
		)
		if err != nil {
			return 0, fmt.Errorf("insert scrobble: %w", err)
		}
		n, _ := result.RowsAffected()
		if n > 0 {
			inserted++
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	return inserted, nil
}

func (r *ScrobbleRepository) GetSyncState(ctx context.Context) (lastPage int, totalImported int, err error) {
	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(last_page, 0), COALESCE(total_imported, 0)
		FROM lastfm_sync_state WHERE id = 1
	`).Scan(&lastPage, &totalImported)
	if err == sql.ErrNoRows {
		return 0, 0, nil
	}
	return
}

func (r *ScrobbleRepository) UpdateSyncState(ctx context.Context, lastPage, totalImported int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT OR REPLACE INTO lastfm_sync_state (id, last_page, total_imported, last_sync_at)
		VALUES (1, ?, ?, CAST(strftime('%s', 'now') AS INTEGER))
	`, lastPage, totalImported)
	return err
}

func (r *ScrobbleRepository) LinkToTracks(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE scrobbles
		SET track_id = (
			SELECT t.id FROM tracks t
			JOIN artists a ON a.id = t.artist_id
			WHERE a.name = scrobbles.artist
			AND t.name = scrobbles.track
			LIMIT 1
		)
		WHERE track_id IS NULL
	`)
	if err != nil {
		return 0, fmt.Errorf("link scrobbles to tracks: %w", err)
	}
	return result.RowsAffected()
}

func (r *ScrobbleRepository) ListByTrack(ctx context.Context, trackID int64, limit int) ([]ScrobbleRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, track_id, artist, artist_mbid, album, album_mbid,
			track, track_mbid, timestamp, loved, url
		FROM scrobbles
		WHERE track_id = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`, trackID, limit)
	if err != nil {
		return nil, fmt.Errorf("list scrobbles by track: %w", err)
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

func (r *ScrobbleRepository) StatsByTrack(ctx context.Context, trackID int64) (count int, first, last *int64, err error) {
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), MIN(timestamp), MAX(timestamp)
		FROM scrobbles WHERE track_id = ?
	`, trackID).Scan(&count, &first, &last)
	return
}

func (r *ScrobbleRepository) StatsByArtist(ctx context.Context, artist string) (count int, first, last *int64, err error) {
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), MIN(timestamp), MAX(timestamp)
		FROM scrobbles WHERE artist = ?
	`, artist).Scan(&count, &first, &last)
	return
}

func (r *ScrobbleRepository) StatsByAlbum(ctx context.Context, albumMBID string) (count int, first, last *int64, err error) {
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), MIN(timestamp), MAX(timestamp)
		FROM scrobbles WHERE album_mbid = ?
	`, albumMBID).Scan(&count, &first, &last)
	return
}

func (r *ScrobbleRepository) GetLatestTimestamp(ctx context.Context) (int64, error) {
	var ts int64
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(timestamp), 0) FROM scrobbles`).Scan(&ts)
	return ts, err
}
