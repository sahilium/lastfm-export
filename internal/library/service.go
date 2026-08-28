package library

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/musiclib/internal/db/repositories"
)

type Service struct {
	db            *sql.DB
	artists       *repositories.ArtistRepository
	albums        *repositories.AlbumRepository
	tracks        *repositories.TrackRepository
	scrobbles     *repositories.ScrobbleRepository
	tags          *repositories.TagRepository
}

func NewService(db *sql.DB) *Service {
	return &Service{
		db:        db,
		artists:   repositories.NewArtistRepository(db),
		albums:    repositories.NewAlbumRepository(db),
		tracks:    repositories.NewTrackRepository(db),
		scrobbles: repositories.NewScrobbleRepository(db),
		tags:      repositories.NewTagRepository(db),
	}
}

func (s *Service) ListArtists(ctx context.Context, page, limit int) ([]Artist, int, error) {
	rows, total, err := s.artists.List(ctx, page, limit)
	if err != nil {
		return nil, 0, err
	}
	artists := make([]Artist, len(rows))
	for i, r := range rows {
		artists[i] = artistFromRow(r)
	}
	return artists, total, nil
}

func (s *Service) GetArtist(ctx context.Context, id int64) (*Artist, error) {
	row, err := s.artists.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	a := artistFromRow(*row)
	return &a, nil
}

func (s *Service) UpdateArtist(ctx context.Context, id int64, note *string, favorite *bool) error {
	return s.artists.Update(ctx, id, note, favorite)
}

func (s *Service) GetArtistAlbums(ctx context.Context, artistID int64) ([]Album, error) {
	rows, err := s.artists.GetAlbums(ctx, artistID)
	if err != nil {
		return nil, err
	}
	albums := make([]Album, len(rows))
	for i, r := range rows {
		albums[i] = albumFromRow(r)
	}
	return albums, nil
}

func (s *Service) GetArtistTracks(ctx context.Context, artistID int64) ([]Track, error) {
	rows, err := s.artists.GetTracks(ctx, artistID)
	if err != nil {
		return nil, err
	}
	tracks := make([]Track, len(rows))
	for i, r := range rows {
		tracks[i] = trackFromRow(r)
	}
	return tracks, nil
}

func (s *Service) ListAlbums(ctx context.Context, page, limit int) ([]Album, int, error) {
	rows, total, err := s.albums.List(ctx, page, limit)
	if err != nil {
		return nil, 0, err
	}
	albums := make([]Album, len(rows))
	for i, r := range rows {
		albums[i] = albumFromRow(r)
	}
	return albums, total, nil
}

func (s *Service) GetAlbum(ctx context.Context, id int64) (*Album, error) {
	row, err := s.albums.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	a := albumFromRow(*row)
	return &a, nil
}

func (s *Service) UpdateAlbum(ctx context.Context, id int64, note *string, favorite *bool) error {
	return s.albums.Update(ctx, id, note, favorite)
}

func (s *Service) GetAlbumTracks(ctx context.Context, albumID int64) ([]Track, error) {
	rows, err := s.albums.GetTracks(ctx, albumID)
	if err != nil {
		return nil, err
	}
	tracks := make([]Track, len(rows))
	for i, r := range rows {
		tracks[i] = trackFromRow(r)
	}
	return tracks, nil
}

func (s *Service) ListTracks(ctx context.Context, page, limit int) ([]Track, int, error) {
	rows, total, err := s.tracks.List(ctx, page, limit)
	if err != nil {
		return nil, 0, err
	}
	tracks := make([]Track, len(rows))
	for i, r := range rows {
		tracks[i] = trackFromRow(r)
	}
	return tracks, total, nil
}

func (s *Service) GetTrack(ctx context.Context, id int64) (*Track, error) {
	row, err := s.tracks.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	t := trackFromRow(*row)
	return &t, nil
}

func (s *Service) UpdateTrack(ctx context.Context, id int64, note *string, favorite *bool) error {
	return s.tracks.Update(ctx, id, note, favorite)
}

func (s *Service) GetTrackHistory(ctx context.Context, trackID int64) ([]Scrobble, error) {
	rows, err := s.scrobbles.ListByTrack(ctx, trackID, 100)
	if err != nil {
		return nil, err
	}
	scrobbles := make([]Scrobble, len(rows))
	for i, r := range rows {
		scrobbles[i] = scrobbleFromRow(r)
	}
	return scrobbles, nil
}

func (s *Service) Search(ctx context.Context, query string) (*SearchResults, error) {
	results := &SearchResults{}

	artistRows, err := s.artists.Search(ctx, query, 10)
	if err != nil {
		slog.Warn("artist search failed", "err", err)
	} else {
		results.Artists.Items = make([]Artist, len(artistRows))
		for i, r := range artistRows {
			results.Artists.Items[i] = artistFromRow(r)
		}
		results.Artists.Total = len(artistRows)
	}

	albumRows, err := s.albums.Search(ctx, query, 10)
	if err != nil {
		slog.Warn("album search failed", "err", err)
	} else {
		results.Albums.Items = make([]Album, len(albumRows))
		for i, r := range albumRows {
			results.Albums.Items[i] = albumFromRow(r)
		}
		results.Albums.Total = len(albumRows)
	}

	trackRows, err := s.tracks.Search(ctx, query, 10)
	if err != nil {
		slog.Warn("track search failed", "err", err)
	} else {
		results.Tracks.Items = make([]Track, len(trackRows))
		for i, r := range trackRows {
			results.Tracks.Items[i] = trackFromRow(r)
		}
		results.Tracks.Total = len(trackRows)
	}

	return results, nil
}

func (s *Service) GetStats(ctx context.Context, entityType string, id int64) (*Stats, error) {
	stats := &Stats{}

	switch entityType {
	case "artist":
		artist, err := s.GetArtist(ctx, id)
		if err != nil {
			return nil, err
		}
		stats.ScrobbleCount = artist.ScrobbleCount
		stats.FirstListened = artist.FirstListened
		stats.LastListened = artist.LastListened

	case "album":
		album, err := s.GetAlbum(ctx, id)
		if err != nil {
			return nil, err
		}
		stats.ScrobbleCount = album.ScrobbleCount
		stats.FirstListened = album.FirstListened
		stats.LastListened = album.LastListened

	case "track":
		track, err := s.GetTrack(ctx, id)
		if err != nil {
			return nil, err
		}
		stats.ScrobbleCount = track.ScrobbleCount
		stats.FirstListened = track.FirstListened
		stats.LastListened = track.LastListened

	default:
		return nil, fmt.Errorf("unknown entity type: %s", entityType)
	}

	return stats, nil
}

func (s *Service) GetTags(ctx context.Context, entityType string, entityID int64) ([]Tag, error) {
	repoTags, err := s.tags.GetForEntity(ctx, entityType, entityID)
	if err != nil {
		return nil, err
	}
	tags := make([]Tag, len(repoTags))
	for i, rt := range repoTags {
		tags[i] = Tag{ID: rt.ID, Name: rt.Name}
	}
	return tags, nil
}

func (s *Service) AddTag(ctx context.Context, entityType string, entityID int64, tagName string) error {
	tagID, err := s.tags.FindOrCreate(ctx, tagName)
	if err != nil {
		return err
	}
	return s.tags.AddToEntity(ctx, tagID, entityType, entityID)
}

func (s *Service) RemoveTag(ctx context.Context, entityType string, entityID int64, tagID int64) error {
	return s.tags.RemoveFromEntity(ctx, tagID, entityType, entityID)
}

func (s *Service) TotalArtists(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM artists`).Scan(&count)
	return count, err
}

func (s *Service) TotalAlbums(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM albums`).Scan(&count)
	return count, err
}

func (s *Service) TotalTracks(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tracks`).Scan(&count)
	return count, err
}

func (s *Service) TotalScrobbles(ctx context.Context) (int, error) {
	return s.scrobbles.Count(ctx)
}

func (s *Service) RebuildFTS(ctx context.Context) error {
	slog.Info("rebuilding FTS indexes")
	if err := s.artists.SyncFTS(ctx); err != nil {
		return fmt.Errorf("rebuild artists fts: %w", err)
	}
	if err := s.albums.SyncFTS(ctx); err != nil {
		return fmt.Errorf("rebuild albums fts: %w", err)
	}
	if err := s.tracks.SyncFTS(ctx); err != nil {
		return fmt.Errorf("rebuild tracks fts: %w", err)
	}
	slog.Info("FTS indexes rebuilt")
	return nil
}

func artistFromRow(r repositories.ArtistRow) Artist {
	a := Artist{
		ID:            r.ID,
		Name:          r.Name,
		Note:          r.Note,
		Favorite:      r.Favorite,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
		ScrobbleCount: r.ScrobbleCount,
		AlbumCount:    r.AlbumCount,
		TrackCount:    r.TrackCount,
	}
	if r.MBID.Valid {
		a.MBID = &r.MBID.String
	}
	if r.FirstListened.Valid {
		v := r.FirstListened.Int64
		a.FirstListened = &v
	}
	if r.LastListened.Valid {
		v := r.LastListened.Int64
		a.LastListened = &v
	}
	return a
}

func albumFromRow(r repositories.AlbumRow) Album {
	al := Album{
		ID:           r.ID,
		ArtistID:     r.ArtistID,
		Name:         r.Name,
		Note:         r.Note,
		Favorite:     r.Favorite,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
		ScrobbleCount: r.ScrobbleCount,
		TrackCount:    r.TrackCount,
	}
	if r.MBID.Valid {
		al.MBID = &r.MBID.String
	}
	if r.ReleaseDate.Valid {
		al.ReleaseDate = &r.ReleaseDate.String
	}
	if r.ArtistName.Valid {
		al.ArtistName = &r.ArtistName.String
	}
	if r.FirstListened.Valid {
		v := r.FirstListened.Int64
		al.FirstListened = &v
	}
	if r.LastListened.Valid {
		v := r.LastListened.Int64
		al.LastListened = &v
	}
	return al
}

func trackFromRow(r repositories.TrackRow) Track {
	tr := Track{
		ID:        r.ID,
		ArtistID:  r.ArtistID,
		Name:      r.Name,
		Note:      r.Note,
		Favorite:  r.Favorite,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
		ScrobbleCount: r.ScrobbleCount,
	}
	if r.AlbumID.Valid {
		tr.AlbumID = &r.AlbumID.Int64
	}
	if r.MBID.Valid {
		tr.MBID = &r.MBID.String
	}
	if r.ArtistName.Valid {
		tr.ArtistName = &r.ArtistName.String
	}
	if r.AlbumName.Valid {
		tr.AlbumName = &r.AlbumName.String
	}
	if r.FirstListened.Valid {
		v := r.FirstListened.Int64
		tr.FirstListened = &v
	}
	if r.LastListened.Valid {
		v := r.LastListened.Int64
		tr.LastListened = &v
	}
	return tr
}

func scrobbleFromRow(r repositories.ScrobbleRow) Scrobble {
	s := Scrobble{
		ID:        r.ID,
		Artist:    r.Artist,
		Track:     r.Track,
		Timestamp: r.Timestamp,
	}
	if r.TrackID.Valid {
		s.TrackID = &r.TrackID.Int64
	}
	if r.ArtistMBID.Valid {
		s.ArtistMBID = &r.ArtistMBID.String
	}
	if r.Album.Valid {
		s.Album = &r.Album.String
	}
	if r.AlbumMBID.Valid {
		s.AlbumMBID = &r.AlbumMBID.String
	}
	if r.TrackMBID.Valid {
		s.TrackMBID = &r.TrackMBID.String
	}
	if r.Loved.Valid {
		v := int(r.Loved.Int64)
		s.Loved = &v
	}
	if r.URL.Valid {
		s.URL = &r.URL.String
	}
	return s
}


