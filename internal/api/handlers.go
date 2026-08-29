package api

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/musiclib/internal/library"
)

type Handler struct {
	svc *library.Service
}

func NewHandler(svc *library.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ListArtists(c *gin.Context) {
	page, limit := parsePagination(c)
	sort := parseSort(c)

	artists, total, err := h.svc.ListArtists(c.Request.Context(), page, limit, sort)
	if err != nil {
		slog.Error("list artists", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list artists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": artists,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *Handler) GetArtist(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	artist, err := h.svc.GetArtist(c.Request.Context(), id)
	if err != nil {
		slog.Error("get artist", "id", id, "err", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "artist not found"})
		return
	}

	c.JSON(http.StatusOK, artist)
}

func (h *Handler) UpdateArtist(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	var body struct {
		Note     *string `json:"note"`
		Favorite *bool   `json:"favorite"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.svc.UpdateArtist(c.Request.Context(), id, body.Note, body.Favorite); err != nil {
		slog.Error("update artist", "id", id, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update artist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) GetArtistAlbums(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	albums, err := h.svc.GetArtistAlbums(c.Request.Context(), id)
	if err != nil {
		slog.Error("get artist albums", "id", id, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get albums"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": albums})
}

func (h *Handler) GetArtistTracks(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	tracks, err := h.svc.GetArtistTracks(c.Request.Context(), id)
	if err != nil {
		slog.Error("get artist tracks", "id", id, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get tracks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": tracks})
}

func (h *Handler) ListAlbums(c *gin.Context) {
	page, limit := parsePagination(c)
	sort := parseSort(c)

	albums, total, err := h.svc.ListAlbums(c.Request.Context(), page, limit, sort)
	if err != nil {
		slog.Error("list albums", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list albums"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": albums,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *Handler) GetAlbum(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	album, err := h.svc.GetAlbum(c.Request.Context(), id)
	if err != nil {
		slog.Error("get album", "id", id, "err", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "album not found"})
		return
	}

	c.JSON(http.StatusOK, album)
}

func (h *Handler) UpdateAlbum(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	var body struct {
		Note     *string `json:"note"`
		Favorite *bool   `json:"favorite"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.svc.UpdateAlbum(c.Request.Context(), id, body.Note, body.Favorite); err != nil {
		slog.Error("update album", "id", id, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update album"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) GetAlbumTracks(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	tracks, err := h.svc.GetAlbumTracks(c.Request.Context(), id)
	if err != nil {
		slog.Error("get album tracks", "id", id, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get tracks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": tracks})
}

func (h *Handler) ListTracks(c *gin.Context) {
	page, limit := parsePagination(c)
	sort := parseSort(c)

	tracks, total, err := h.svc.ListTracks(c.Request.Context(), page, limit, sort)
	if err != nil {
		slog.Error("list tracks", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tracks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": tracks,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *Handler) GetTrack(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	track, err := h.svc.GetTrack(c.Request.Context(), id)
	if err != nil {
		slog.Error("get track", "id", id, "err", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "track not found"})
		return
	}

	c.JSON(http.StatusOK, track)
}

func (h *Handler) UpdateTrack(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	var body struct {
		Note     *string `json:"note"`
		Favorite *bool   `json:"favorite"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.svc.UpdateTrack(c.Request.Context(), id, body.Note, body.Favorite); err != nil {
		slog.Error("update track", "id", id, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update track"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) GetTrackHistory(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	history, err := h.svc.GetTrackHistory(c.Request.Context(), id)
	if err != nil {
		slog.Error("get track history", "id", id, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": history})
}

func (h *Handler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	results, err := h.svc.Search(c.Request.Context(), q)
	if err != nil {
		slog.Error("search", "q", q, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (h *Handler) GetStats(c *gin.Context) {
	entityType := c.Param("type")
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	stats, err := h.svc.GetStats(c.Request.Context(), entityType, id)
	if err != nil {
		slog.Error("get stats", "type", entityType, "id", id, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) GetTags(c *gin.Context) {
	entityType := c.Param("type")
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	tags, err := h.svc.GetTags(c.Request.Context(), entityType, id)
	if err != nil {
		slog.Error("get tags", "type", entityType, "id", id, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get tags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": tags})
}

func (h *Handler) AddTag(c *gin.Context) {
	entityType := c.Param("type")
	id, err := parseID(c, "id")
	if err != nil {
		return
	}

	var body struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag name is required"})
		return
	}

	if err := h.svc.AddTag(c.Request.Context(), entityType, id, body.Name); err != nil {
		slog.Error("add tag", "type", entityType, "id", id, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add tag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) RemoveTag(c *gin.Context) {
	entityType := c.Param("type")
	id, err := parseID(c, "id")
	if err != nil {
		return
	}
	tagID, err := parseID(c, "tagId")
	if err != nil {
		return
	}

	if err := h.svc.RemoveTag(c.Request.Context(), entityType, id, tagID); err != nil {
		slog.Error("remove tag", "type", entityType, "id", id, "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove tag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) Summary(c *gin.Context) {
	ctx := c.Request.Context()

	artists, _ := h.svc.TotalArtists(ctx)
	albums, _ := h.svc.TotalAlbums(ctx)
	tracks, _ := h.svc.TotalTracks(ctx)
	scrobbles, _ := h.svc.TotalScrobbles(ctx)

	c.JSON(http.StatusOK, gin.H{
		"artists":   artists,
		"albums":    albums,
		"tracks":    tracks,
		"scrobbles": scrobbles,
	})
}

func parseID(c *gin.Context, param string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(param), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + param})
		return 0, err
	}
	return id, nil
}

func parsePagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}

	return page, limit
}

var validSorts = map[string]bool{
	"name_asc": true, "name_desc": true,
	"scrobbles_asc": true, "scrobbles_desc": true,
	"recent_asc": true, "recent_desc": true,
}

func parseSort(c *gin.Context) string {
	sort := c.DefaultQuery("sort", "name_asc")
	if !validSorts[sort] {
		return "name_asc"
	}
	return sort
}
