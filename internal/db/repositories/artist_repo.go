package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ArtistRow struct {
	ID            int64
	Name          string
	MBID          sql.NullString
	Note          string
	Favorite      bool
	CreatedAt     int64
	UpdatedAt     int64
	ScrobbleCount int
	AlbumCount    int
	TrackCount    int
	FirstListened sql.NullInt64
	LastListened  sql.NullInt64
}

type ArtistRepository struct {
	db *sql.DB
}

func NewArtistRepository(db *sql.DB) *ArtistRepository {
	return &ArtistRepository{db: db}
}

func artistSortClause(sort string) string {
	switch sort {
	case "name_desc":
		return "a.name DESC"
	case "scrobbles_desc":
		return "COALESCE(s.scrobble_count, 0) DESC, a.name ASC"
	case "scrobbles_asc":
		return "COALESCE(s.scrobble_count, 0) ASC, a.name ASC"
	case "recent_desc":
		return "s.last_listened DESC NULLS LAST, a.name ASC"
	case "recent_asc":
		return "s.last_listened ASC NULLS FIRST, a.name ASC"
	default:
		return "a.name ASC"
	}
}

const artistListQuery = `
	SELECT
		a.id, a.name, a.mbid, a.note, a.favorite,
		a.created_at, a.updated_at,
		COALESCE(s.scrobble_count, 0),
		COALESCE(al.album_count, 0),
		COALESCE(t.track_count, 0),
		s.first_listened,
		s.last_listened
	FROM artists a
	LEFT JOIN (
		SELECT artist, COUNT(*) as scrobble_count,
			MIN(timestamp) as first_listened,
			MAX(timestamp) as last_listened
		FROM scrobbles
		GROUP BY artist
	) s ON s.artist = a.name
	LEFT JOIN (
		SELECT artist_id, COUNT(*) as album_count
		FROM albums
		GROUP BY artist_id
	) al ON al.artist_id = a.id
	LEFT JOIN (
		SELECT artist_id, COUNT(*) as track_count
		FROM tracks
		GROUP BY artist_id
	) t ON t.artist_id = a.id
`

func (r *ArtistRepository) List(ctx context.Context, page, limit int, sort string) ([]ArtistRow, int, error) {
	offset := (page - 1) * limit

	var total int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM artists`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count artists: %w", err)
	}

	query := artistListQuery + ` ORDER BY ` + artistSortClause(sort) + ` LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list artists: %w", err)
	}
	defer rows.Close()

	var artists []ArtistRow
	for rows.Next() {
		var a ArtistRow
		err := rows.Scan(
			&a.ID, &a.Name, &a.MBID, &a.Note, &a.Favorite,
			&a.CreatedAt, &a.UpdatedAt,
			&a.ScrobbleCount, &a.AlbumCount, &a.TrackCount,
			&a.FirstListened, &a.LastListened,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan artist: %w", err)
		}
		artists = append(artists, a)
	}

	return artists, total, rows.Err()
}

func (r *ArtistRepository) GetByID(ctx context.Context, id int64) (*ArtistRow, error) {
	var a ArtistRow
	err := r.db.QueryRowContext(ctx, artistListQuery+` WHERE a.id = ?`, id).Scan(
		&a.ID, &a.Name, &a.MBID, &a.Note, &a.Favorite,
		&a.CreatedAt, &a.UpdatedAt,
		&a.ScrobbleCount, &a.AlbumCount, &a.TrackCount,
		&a.FirstListened, &a.LastListened,
	)
	if err != nil {
		return nil, fmt.Errorf("get artist: %w", err)
	}
	return &a, nil
}

func (r *ArtistRepository) GetByMBID(ctx context.Context, mbid string) (*ArtistRow, error) {
	var a ArtistRow
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, mbid, note, favorite, created_at, updated_at
		FROM artists WHERE mbid = ?
	`, mbid).Scan(&a.ID, &a.Name, &a.MBID, &a.Note, &a.Favorite, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get artist by mbid: %w", err)
	}
	return &a, nil
}

func (r *ArtistRepository) FindOrCreate(ctx context.Context, name string, mbid *string, now time.Time) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `SELECT id FROM artists WHERE name = ?`, name).Scan(&id)
	if err == nil {
		if mbid != nil {
			_, _ = r.db.ExecContext(ctx, `UPDATE artists SET mbid = ? WHERE id = ? AND mbid IS NULL`, *mbid, id)
		}
		return id, nil
	}

	var mbidVal sql.NullString
	if mbid != nil {
		mbidVal = sql.NullString{String: *mbid, Valid: true}
	}

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO artists (name, mbid, note, favorite, created_at, updated_at)
		VALUES (?, ?, '', 0, ?, ?)
	`, name, mbidVal, now, now)
	if err != nil {
		return 0, fmt.Errorf("insert artist: %w", err)
	}

	return result.LastInsertId()
}

func (r *ArtistRepository) Update(ctx context.Context, id int64, note *string, favorite *bool) error {
	if note != nil {
		if _, err := r.db.ExecContext(ctx, `UPDATE artists SET note = ?, updated_at = ? WHERE id = ?`, *note, time.Now().Unix(), id); err != nil {
			return fmt.Errorf("update artist note: %w", err)
		}
	}
	if favorite != nil {
		if _, err := r.db.ExecContext(ctx, `UPDATE artists SET favorite = ?, updated_at = ? WHERE id = ?`, *favorite, time.Now().Unix(), id); err != nil {
			return fmt.Errorf("update artist favorite: %w", err)
		}
	}
	return nil
}

func (r *ArtistRepository) Search(ctx context.Context, query string, limit int) ([]ArtistRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.id, a.name, a.mbid, a.note, a.favorite, a.created_at, a.updated_at,
			COALESCE(s.scrobble_count, 0)
		FROM artists a
		JOIN artists_fts fts ON fts.rowid = a.id
		LEFT JOIN (
			SELECT artist, COUNT(*) as scrobble_count
			FROM scrobbles
			GROUP BY artist
		) s ON s.artist = a.name
		WHERE artists_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search artists: %w", err)
	}
	defer rows.Close()

	var artists []ArtistRow
	for rows.Next() {
		var a ArtistRow
		if err := rows.Scan(&a.ID, &a.Name, &a.MBID, &a.Note, &a.Favorite, &a.CreatedAt, &a.UpdatedAt, &a.ScrobbleCount); err != nil {
			return nil, fmt.Errorf("scan artist: %w", err)
		}
		artists = append(artists, a)
	}
	return artists, rows.Err()
}

func (r *ArtistRepository) GetAlbums(ctx context.Context, artistID int64) ([]AlbumRow, error) {
	rows, err := r.db.QueryContext(ctx, `
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
		WHERE al.artist_id = ?
		ORDER BY al.name ASC
	`, artistID)
	if err != nil {
		return nil, fmt.Errorf("get artist albums: %w", err)
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
			return nil, fmt.Errorf("scan album: %w", err)
		}
		albums = append(albums, al)
	}
	return albums, rows.Err()
}

func (r *ArtistRepository) GetTracks(ctx context.Context, artistID int64) ([]TrackRow, error) {
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
		WHERE t.artist_id = ?
		ORDER BY t.name ASC
	`, artistID)
	if err != nil {
		return nil, fmt.Errorf("get artist tracks: %w", err)
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

func (r *ArtistRepository) SyncFTS(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO artists_fts(artists_fts) VALUES('rebuild')
	`)
	return err
}
