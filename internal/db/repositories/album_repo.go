package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type AlbumRow struct {
	ID           int64
	ArtistID     int64
	Name         string
	MBID         sql.NullString
	ReleaseDate  sql.NullString
	Note         string
	Favorite     bool
	CreatedAt    int64
	UpdatedAt    int64
	ArtistName   sql.NullString
	ScrobbleCount int
	TrackCount    int
	FirstListened sql.NullInt64
	LastListened  sql.NullInt64
}

type AlbumRepository struct {
	db *sql.DB
}

func NewAlbumRepository(db *sql.DB) *AlbumRepository {
	return &AlbumRepository{db: db}
}

func albumSortClause(sort string) string {
	switch sort {
	case "name_desc":
		return "al.name DESC"
	case "scrobbles_desc":
		return "COALESCE(s.scrobble_count, 0) DESC, al.name ASC"
	case "scrobbles_asc":
		return "COALESCE(s.scrobble_count, 0) ASC, al.name ASC"
	case "recent_desc":
		return "s.last_listened DESC NULLS LAST, al.name ASC"
	case "recent_asc":
		return "s.last_listened ASC NULLS FIRST, al.name ASC"
	default:
		return "al.name ASC"
	}
}

const albumListQuery = `
	SELECT
		al.id, al.artist_id, al.name, al.mbid, al.release_date,
		al.note, al.favorite, al.created_at, al.updated_at,
		a.name,
		COALESCE(s.scrobble_count, 0),
		COALESCE(t.track_count, 0),
		s.first_listened,
		s.last_listened
	FROM albums al
	JOIN artists a ON a.id = al.artist_id
	LEFT JOIN (
		SELECT sc.album, sc.artist, COUNT(*) as scrobble_count,
			MIN(sc.timestamp) as first_listened,
			MAX(sc.timestamp) as last_listened
		FROM scrobbles sc
		WHERE sc.album IS NOT NULL AND sc.album != ''
		GROUP BY sc.artist, sc.album
	) s ON s.artist = a.name AND s.album = al.name
	LEFT JOIN (
		SELECT album_id, COUNT(*) as track_count
		FROM tracks
		WHERE album_id IS NOT NULL
		GROUP BY album_id
	) t ON t.album_id = al.id
`

func (r *AlbumRepository) List(ctx context.Context, page, limit int, sort string) ([]AlbumRow, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM albums`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count albums: %w", err)
	}

	query := albumListQuery + ` ORDER BY ` + albumSortClause(sort) + ` LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list albums: %w", err)
	}
	defer rows.Close()

	var albums []AlbumRow
	for rows.Next() {
		var al AlbumRow
		if err := rows.Scan(
			&al.ID, &al.ArtistID, &al.Name, &al.MBID, &al.ReleaseDate,
			&al.Note, &al.Favorite, &al.CreatedAt, &al.UpdatedAt,
			&al.ArtistName, &al.ScrobbleCount, &al.TrackCount,
			&al.FirstListened, &al.LastListened,
		); err != nil {
			return nil, 0, fmt.Errorf("scan album: %w", err)
		}
		albums = append(albums, al)
	}

	return albums, total, rows.Err()
}

func (r *AlbumRepository) GetByID(ctx context.Context, id int64) (*AlbumRow, error) {
	var al AlbumRow
	err := r.db.QueryRowContext(ctx, albumListQuery+` WHERE al.id = ?`, id).Scan(
		&al.ID, &al.ArtistID, &al.Name, &al.MBID, &al.ReleaseDate,
		&al.Note, &al.Favorite, &al.CreatedAt, &al.UpdatedAt,
		&al.ArtistName, &al.ScrobbleCount, &al.TrackCount,
		&al.FirstListened, &al.LastListened,
	)
	if err != nil {
		return nil, fmt.Errorf("get album: %w", err)
	}
	return &al, nil
}

func (r *AlbumRepository) FindOrCreate(ctx context.Context, artistID int64, name string, mbid *string, now int64) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `SELECT id FROM albums WHERE artist_id = ? AND name = ?`, artistID, name).Scan(&id)
	if err == nil {
		if mbid != nil {
			_, _ = r.db.ExecContext(ctx, `UPDATE albums SET mbid = ? WHERE id = ? AND mbid IS NULL`, *mbid, id)
		}
		return id, nil
	}

	var mbidVal sql.NullString
	if mbid != nil {
		mbidVal = sql.NullString{String: *mbid, Valid: true}
	}

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO albums (artist_id, name, mbid, note, favorite, created_at, updated_at)
		VALUES (?, ?, ?, '', 0, ?, ?)
	`, artistID, name, mbidVal, now, now)
	if err != nil {
		return 0, fmt.Errorf("insert album: %w", err)
	}

	return result.LastInsertId()
}

func (r *AlbumRepository) Update(ctx context.Context, id int64, note *string, favorite *bool) error {
	if note != nil {
		if _, err := r.db.ExecContext(ctx, `UPDATE albums SET note = ?, updated_at = ? WHERE id = ?`, *note, time.Now().Unix(), id); err != nil {
			return fmt.Errorf("update album note: %w", err)
		}
	}
	if favorite != nil {
		if _, err := r.db.ExecContext(ctx, `UPDATE albums SET favorite = ?, updated_at = ? WHERE id = ?`, *favorite, time.Now().Unix(), id); err != nil {
			return fmt.Errorf("update album favorite: %w", err)
		}
	}
	return nil
}

func (r *AlbumRepository) Search(ctx context.Context, query string, limit int) ([]AlbumRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT al.id, al.artist_id, al.name, al.mbid, al.release_date,
			al.note, al.favorite, al.created_at, al.updated_at,
			a.name,
			COALESCE(s.scrobble_count, 0),
			COALESCE(t.track_count, 0),
			s.first_listened,
			s.last_listened
		FROM albums al
		JOIN artists a ON a.id = al.artist_id
		JOIN albums_fts fts ON fts.rowid = al.id
		LEFT JOIN (
			SELECT sc.album, sc.artist, COUNT(*) as scrobble_count,
				MIN(sc.timestamp) as first_listened,
				MAX(sc.timestamp) as last_listened
			FROM scrobbles sc
			WHERE sc.album IS NOT NULL AND sc.album != ''
			GROUP BY sc.artist, sc.album
		) s ON s.artist = a.name AND s.album = al.name
		LEFT JOIN (
			SELECT album_id, COUNT(*) as track_count
			FROM tracks
			WHERE album_id IS NOT NULL
			GROUP BY album_id
		) t ON t.album_id = al.id
		WHERE albums_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search albums: %w", err)
	}
	defer rows.Close()

	var albums []AlbumRow
	for rows.Next() {
		var al AlbumRow
		if err := rows.Scan(&al.ID, &al.ArtistID, &al.Name, &al.MBID, &al.ReleaseDate,
			&al.Note, &al.Favorite, &al.CreatedAt, &al.UpdatedAt,
			&al.ArtistName, &al.ScrobbleCount, &al.TrackCount,
			&al.FirstListened, &al.LastListened); err != nil {
			return nil, fmt.Errorf("scan album: %w", err)
		}
		albums = append(albums, al)
	}
	return albums, rows.Err()
}

func (r *AlbumRepository) GetTracks(ctx context.Context, albumID int64) ([]TrackRow, error) {
	rows, err := r.db.QueryContext(ctx, `
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
		WHERE t.album_id = ?
		ORDER BY t.name ASC
	`, albumID)
	if err != nil {
		return nil, fmt.Errorf("get album tracks: %w", err)
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
			return nil, fmt.Errorf("scan track: %w", err)
		}
		tracks = append(tracks, tr)
	}
	return tracks, rows.Err()
}

func (r *AlbumRepository) SyncFTS(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO albums_fts(albums_fts) VALUES('rebuild')
	`)
	return err
}
