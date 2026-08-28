package main

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/musiclib/internal/api"
	"github.com/musiclib/internal/config"
	"github.com/musiclib/internal/db"
	"github.com/musiclib/internal/db/migrations"
	"github.com/musiclib/internal/library"
	"github.com/musiclib/web"
)

func main() {
	if len(os.Args) < 2 {
		serve()
		return
	}

	switch os.Args[1] {
	case "serve":
		serve()
	case "migrate":
		runMigrate()
	case "import":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "usage: musiclib import <source-db-path>\n")
			os.Exit(1)
		}
		runImport(os.Args[2])
	case "stats":
		runStats()
	case "fts":
		runFTSRebuild()
	default:
		fmt.Fprintf(os.Stderr, "usage: musiclib [serve|migrate|import <source>|stats|fts]\n")
		os.Exit(1)
	}
}

func serve() {
	cfg := config.Load()
	setupLogging()

	slog.Info("starting musiclib", "address", cfg.ServerAddress)

	database, err := db.Open(db.Options{
		DatabasePath: cfg.DatabasePath,
		TursoURL:     cfg.TursoURL,
		TursoToken:   cfg.TursoAuthToken,
	})
	if err != nil {
		slog.Error("failed to open database", "err", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := migrations.Run(database); err != nil {
		slog.Error("failed to run migrations", "err", err)
		os.Exit(1)
	}

	svc := library.NewService(database)

	lastfmHandler := api.NewLastfmSyncHandler(database, cfg.LastfmAPIKey, cfg.LastfmUsername)
	router := api.NewRouter(svc, lastfmHandler)

	// Serve embedded frontend
	embedFrontend(router)

	slog.Info("server starting", "address", cfg.ServerAddress)
	if err := router.Run(cfg.ServerAddress); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func embedFrontend(r *gin.Engine) {
	subFS, err := fs.Sub(web.FS, "dist")
	if err != nil {
		slog.Warn("failed to get frontend sub filesystem, falling back to no frontend", "err", err)
		return
	}

	indexHTML, err := fs.ReadFile(subFS, "index.html")
	if err != nil {
		slog.Warn("failed to read index.html", "err", err)
		return
	}

	fileServer := http.FileServer(http.FS(subFS))

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Try to serve the static file directly
		if path != "/" && path != "" {
			cleaned := path[1:]
			if f, err := subFS.(fs.ReadFileFS).Open(cleaned); err == nil {
				f.Close()
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
		}

		// SPA fallback: serve index.html
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})

	slog.Info("embedded frontend loaded")
}

func runMigrate() {
	cfg := config.Load()
	setupLogging()

	database, err := db.Open(db.Options{
		DatabasePath: cfg.DatabasePath,
		TursoURL:     cfg.TursoURL,
		TursoToken:   cfg.TursoAuthToken,
	})
	if err != nil {
		slog.Error("failed to open database", "err", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := migrations.Run(database); err != nil {
		slog.Error("migration failed", "err", err)
		os.Exit(1)
	}

	slog.Info("migrations complete")
}

func runStats() {
	cfg := config.Load()
	setupLogging()

	database, err := db.Open(db.Options{
		DatabasePath: cfg.DatabasePath,
		TursoURL:     cfg.TursoURL,
		TursoToken:   cfg.TursoAuthToken,
	})
	if err != nil {
		slog.Error("failed to open database", "err", err)
		os.Exit(1)
	}
	defer database.Close()

	var count int
	database.QueryRow(`SELECT COUNT(*) FROM scrobbles`).Scan(&count)
	fmt.Printf("Scrobbles: %d\n", count)

	database.QueryRow(`SELECT COUNT(*) FROM artists`).Scan(&count)
	fmt.Printf("Artists: %d\n", count)

	database.QueryRow(`SELECT COUNT(*) FROM albums`).Scan(&count)
	fmt.Printf("Albums: %d\n", count)

	database.QueryRow(`SELECT COUNT(*) FROM tracks`).Scan(&count)
	fmt.Printf("Tracks: %d\n", count)
}

func runFTSRebuild() {
	cfg := config.Load()
	setupLogging()

	database, err := db.Open(db.Options{
		DatabasePath: cfg.DatabasePath,
		TursoURL:     cfg.TursoURL,
		TursoToken:   cfg.TursoAuthToken,
	})
	if err != nil {
		slog.Error("failed to open database", "err", err)
		os.Exit(1)
	}
	defer database.Close()

	svc := library.NewService(database)
	if err := svc.RebuildFTS(context.Background()); err != nil {
		slog.Error("fts rebuild failed", "err", err)
		os.Exit(1)
	}

	slog.Info("fts indexes rebuilt")
}

func setupLogging() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}
