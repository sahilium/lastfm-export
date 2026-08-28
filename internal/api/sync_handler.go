package api

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/musiclib/internal/db/repositories"
	"github.com/musiclib/internal/lastfm"
)

type LastfmSyncHandler struct {
	db         *sql.DB
	apiKey     string
	username   string
}

func NewLastfmSyncHandler(db *sql.DB, apiKey, username string) *LastfmSyncHandler {
	return &LastfmSyncHandler{
		db:       db,
		apiKey:   apiKey,
		username: username,
	}
}

func (h *LastfmSyncHandler) Start(c *gin.Context) {
	if h.apiKey == "" || h.username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "LASTFM_API_KEY and LASTFM_USERNAME must be configured"})
		return
	}

	slog.Info("starting last.fm sync", "username", h.username)

	go h.runSync()

	c.JSON(http.StatusOK, gin.H{
		"status":  "started",
		"message": "Last.fm sync started in background",
	})
}

func (h *LastfmSyncHandler) runSync() {
	ctx := context.Background()
	start := time.Now()
	slog.Info("last.fm sync started")

	client := lastfm.NewClient(h.apiKey, h.username)
	sr := repositories.NewScrobbleRepository(h.db)

	lastPage, totalImported, err := sr.GetSyncState(ctx)
	if err != nil {
		slog.Error("failed to get sync state", "err", err)
		return
	}

	if lastPage > 0 {
		slog.Info("resuming sync", "from_page", lastPage, "already_imported", totalImported)
	}

	page := lastPage + 1
	if page < 1 {
		page = 1
	}

	firstPage, err := client.GetRecentTracks(page)
	if err != nil {
		slog.Error("failed to get first page", "err", err)
		return
	}

	slog.Info("last.fm history", "total_scrobbles", firstPage.Total, "total_pages", firstPage.TotalPages)

	inserted := totalImported

	for p := page; p <= firstPage.TotalPages; p++ {
		if p > page {
			client.RateLimit()
		}

		var result *lastfm.PageResult
		if p == page {
			result = firstPage
		} else {
			result, err = client.GetRecentTracks(p)
			if err != nil {
				slog.Error("failed to get page", "page", p, "err", err)
				continue
			}
		}

		scrobbles := make([]repositories.ScrobbleRow, 0, len(result.Tracks))

		for _, t := range result.Tracks {
			if t.Date == nil || t.Date.UTS == "" {
				continue
			}

			var timestamp int64
			fmt.Sscanf(t.Date.UTS, "%d", &timestamp)

			if timestamp == 0 {
				continue
			}

			s := repositories.ScrobbleRow{
				Artist: t.Artist.Text,
				Track:  t.Name,
				Timestamp: timestamp,
			}

			if t.Artist.MBID != "" {
				s.ArtistMBID = sql.NullString{String: t.Artist.MBID, Valid: true}
			}
			if t.Album.Text != "" {
				s.Album = sql.NullString{String: t.Album.Text, Valid: true}
			}
			if t.Album.MBID != "" {
				s.AlbumMBID = sql.NullString{String: t.Album.MBID, Valid: true}
			}
			if t.MBID != "" {
				s.TrackMBID = sql.NullString{String: t.MBID, Valid: true}
			}
			if t.URL != "" {
				s.URL = sql.NullString{String: t.URL, Valid: true}
			}

			scrobbles = append(scrobbles, s)
		}

		n, err := sr.InsertBatch(ctx, scrobbles)
		if err != nil {
			slog.Error("failed to insert batch", "page", p, "err", err)
			continue
		}

		inserted += n

		if err := sr.UpdateSyncState(ctx, p, inserted); err != nil {
			slog.Warn("failed to update sync state", "err", err)
		}

		if p%10 == 0 || p == firstPage.TotalPages {
			slog.Info("sync progress",
				"page", p,
				"total_pages", firstPage.TotalPages,
				"new_this_page", n,
				"total_new", inserted,
			)
		}
	}

	elapsed := time.Since(start)
	slog.Info("last.fm sync complete",
		"duration", elapsed,
		"total_imported", inserted,
	)
}


