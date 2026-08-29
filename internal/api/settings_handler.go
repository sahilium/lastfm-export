package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/musiclib/internal/db/repositories"
)

type SettingsHandler struct {
	repo      *repositories.SettingsRepository
	localMode bool
}

func NewSettingsHandler(repo *repositories.SettingsRepository, localMode bool) *SettingsHandler {
	return &SettingsHandler{repo: repo, localMode: localMode}
}

func (h *SettingsHandler) Get(c *gin.Context) {
	tursoURL, _ := h.repo.Get(c.Request.Context(), "turso_url")
	tursoToken, _ := h.repo.Get(c.Request.Context(), "turso_token")

	tokenMasked := ""
	if len(tursoToken) > 4 {
		tokenMasked = tursoToken[:2] + "****" + tursoToken[len(tursoToken)-2:]
	} else if tursoToken != "" {
		tokenMasked = "****"
	}

	mode := "turso"
	if h.localMode {
		mode = "local"
	} else if tursoURL == "" {
		mode = "unconfigured"
	}

	c.JSON(http.StatusOK, gin.H{
		"mode":          mode,
		"turso_url":     tursoURL,
		"has_token":     tursoToken != "",
		"turso_token":   tokenMasked,
		"local_mode":    h.localMode,
	})
}

func (h *SettingsHandler) Update(c *gin.Context) {
	var body struct {
		TursoURL   *string `json:"turso_url"`
		TursoToken *string `json:"turso_token"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	ctx := c.Request.Context()

	if body.TursoURL != nil {
		if err := h.repo.Set(ctx, "turso_url", *body.TursoURL); err != nil {
			slog.Error("set turso_url", "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save setting"})
			return
		}
	}

	if body.TursoToken != nil {
		if err := h.repo.Set(ctx, "turso_token", *body.TursoToken); err != nil {
			slog.Error("set turso_token", "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save setting"})
			return
		}
	}

	slog.Info("settings updated")
	c.JSON(http.StatusOK, gin.H{"ok": true, "restart_required": true})
}
