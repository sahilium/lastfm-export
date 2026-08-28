package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabasePath     string
	ServerAddress    string
	LastfmAPIKey     string
	LastfmUsername   string
	MusicBrainzAgent string
	TursoURL         string
	TursoAuthToken   string
}

func Load() *Config {
	return &Config{
		DatabasePath:     envOrDefault("DATABASE_PATH", "musiclib.db"),
		ServerAddress:    envOrDefault("SERVER_ADDRESS", ":8080"),
		LastfmAPIKey:     os.Getenv("LASTFM_API_KEY"),
		LastfmUsername:   os.Getenv("LASTFM_USERNAME"),
		MusicBrainzAgent: envOrDefault("MUSICBRAINZ_USER_AGENT", "musiclib/1.0 (personal music library)"),
		TursoURL:         os.Getenv("TURSO_DATABASE_URL"),
		TursoAuthToken:   os.Getenv("TURSO_AUTH_TOKEN"),
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envIntOrDefault(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
