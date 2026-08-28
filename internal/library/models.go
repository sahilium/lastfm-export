package library

type Artist struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	MBID      *string `json:"mbid"`
	Note      string  `json:"note"`
	Favorite  bool    `json:"favorite"`
	CreatedAt int64   `json:"created_at"`
	UpdatedAt int64   `json:"updated_at"`

	ScrobbleCount int    `json:"scrobble_count"`
	AlbumCount    int    `json:"album_count"`
	TrackCount    int    `json:"track_count"`
	FirstListened *int64 `json:"first_listened"`
	LastListened  *int64 `json:"last_listened"`
}

type Album struct {
	ID          int64   `json:"id"`
	ArtistID    int64   `json:"artist_id"`
	Name        string  `json:"name"`
	MBID        *string `json:"mbid"`
	ReleaseDate *string `json:"release_date"`
	Note        string  `json:"note"`
	Favorite    bool    `json:"favorite"`
	CreatedAt   int64   `json:"created_at"`
	UpdatedAt   int64   `json:"updated_at"`

	ArtistName    *string `json:"artist_name"`
	ScrobbleCount int     `json:"scrobble_count"`
	TrackCount    int     `json:"track_count"`
	FirstListened *int64  `json:"first_listened"`
	LastListened  *int64  `json:"last_listened"`
}

type Track struct {
	ID        int64   `json:"id"`
	ArtistID  int64   `json:"artist_id"`
	AlbumID   *int64  `json:"album_id"`
	Name      string  `json:"name"`
	MBID      *string `json:"mbid"`
	Note      string  `json:"note"`
	Favorite  bool    `json:"favorite"`
	CreatedAt int64   `json:"created_at"`
	UpdatedAt int64   `json:"updated_at"`

	ArtistName    *string `json:"artist_name"`
	AlbumName     *string `json:"album_name"`
	ScrobbleCount int     `json:"scrobble_count"`
	FirstListened *int64  `json:"first_listened"`
	LastListened  *int64  `json:"last_listened"`
}

type Scrobble struct {
	ID         int64   `json:"id"`
	TrackID    *int64  `json:"track_id"`
	Artist     string  `json:"artist"`
	ArtistMBID *string `json:"artist_mbid"`
	Album      *string `json:"album"`
	AlbumMBID  *string `json:"album_mbid"`
	Track      string  `json:"track"`
	TrackMBID  *string `json:"track_mbid"`
	Timestamp  int64   `json:"timestamp"`
	Loved      *int    `json:"loved"`
	URL        *string `json:"url"`
}

type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type EntityTag struct {
	ID         int64  `json:"id"`
	TagID      int64  `json:"tag_id"`
	EntityType string `json:"entity_type"`
	EntityID   int64  `json:"entity_id"`
	CreatedAt  int64  `json:"created_at"`
	TagName    string `json:"tag_name"`
}

type PaginatedResult[T any] struct {
	Items []T   `json:"items"`
	Total int   `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

type Stats struct {
	ScrobbleCount int    `json:"scrobble_count"`
	FirstListened *int64 `json:"first_listened"`
	LastListened  *int64 `json:"last_listened"`
}

type SearchResults struct {
	Artists PaginatedResult[Artist] `json:"artists"`
	Albums  PaginatedResult[Album]  `json:"albums"`
	Tracks  PaginatedResult[Track]  `json:"tracks"`
}
