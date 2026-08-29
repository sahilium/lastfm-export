package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/musiclib/internal/api"
	"github.com/musiclib/internal/config"
	"github.com/musiclib/internal/curation"
	"github.com/musiclib/internal/db"
	"github.com/musiclib/internal/db/migrations"
	"github.com/musiclib/internal/db/repositories"
	"github.com/musiclib/internal/library"
	"github.com/musiclib/web"
)

const settingsDBPath = "musiclib-settings.db"

func main() {
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		serveCmd(false, "")
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "serve":
		fs := flag.NewFlagSet("serve", flag.ExitOnError)
		dbPath := fs.String("db", "", "path to local SQLite database")
		_ = fs.Parse(os.Args[2:])
		serveCmd(*dbPath != "", *dbPath)
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
		fmt.Fprintf(os.Stderr, "usage: musiclib [serve [--db <path>]|migrate|import <source>|stats|fts]\n")
		os.Exit(1)
	}
}

func serveCmd(localMode bool, localDBPath string) {
	cfg := config.Load()
	setupLogging()

	slog.Info("starting musiclib", "address", cfg.ServerAddress)

	settingsDB, err := db.Open(db.Options{DatabasePath: settingsDBPath})
	if err != nil {
		slog.Error("failed to open settings database", "err", err)
		os.Exit(1)
	}
	defer settingsDB.Close()

	if err := migrations.Run(settingsDB); err != nil {
		slog.Error("failed to run settings migrations", "err", err)
		os.Exit(1)
	}

	settingsRepo := repositories.NewSettingsRepository(settingsDB)

	var dataDB *sql.DB

	if localMode {
		slog.Info("local mode", "path", localDBPath)
		w, err := db.Open(db.Options{DatabasePath: localDBPath})
		if err != nil {
			slog.Error("failed to open local database", "err", err)
			os.Exit(1)
		}
		defer w.Close()
		dataDB = w
	} else {
		tursoURL, _ := settingsRepo.Get(context.Background(), "turso_url")
		tursoToken, _ := settingsRepo.Get(context.Background(), "turso_token")

		if tursoURL == "" {
			slog.Warn("no turso configured — server started but data features unavailable. configure turso in /settings and restart.")
			dataDB = nil
		} else {
			slog.Info("connecting to turso")
			w, err := db.Open(db.Options{TursoURL: tursoURL, TursoToken: tursoToken})
			if err != nil {
				slog.Error("failed to connect to turso", "err", err)
				dataDB = nil
			} else {
				defer w.Close()
				if err := migrations.Run(w); err != nil {
					slog.Error("failed to run turso migrations", "err", err)
					dataDB = nil
				} else {
					dataDB = w
				}
			}
		}
	}

	var svc *library.Service
	var curationSvc *curation.Service
	var lastfmHandler *api.LastfmSyncHandler

	if dataDB != nil {
		svc = library.NewService(dataDB)
		collectionRepo := repositories.NewCollectionRepository(dataDB)
		curationSvc = curation.NewService(collectionRepo)
		lastfmHandler = api.NewLastfmSyncHandler(dataDB, cfg.LastfmAPIKey, cfg.LastfmUsername)
	}

	router := api.NewRouter(svc, curationSvc, settingsRepo, lastfmHandler, dataDB != nil)
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

		if path != "/" && path != "" {
			cleaned := path[1:]
			if f, err := subFS.(fs.ReadFileFS).Open(cleaned); err == nil {
				f.Close()
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})

	slog.Info("embedded frontend loaded")
}

func runMigrate() {
	setupLogging()

	database, err := db.Open(db.Options{DatabasePath: "musiclib.db"})
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
	setupLogging()

	database, err := db.Open(db.Options{DatabasePath: "musiclib.db"})
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
	setupLogging()

	database, err := db.Open(db.Options{DatabasePath: "musiclib.db"})
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
