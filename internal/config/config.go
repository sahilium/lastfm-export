package config

import (
	"os"
)

type Config struct {
	ServerAddress    string
	LastfmAPIKey     string
	LastfmUsername   string
	MusicBrainzAgent string
}

func Load() *Config {
	return &Config{
		ServerAddress:    envOrDefault("SERVER_ADDRESS", ":8080"),
		LastfmAPIKey:     os.Getenv("LASTFM_API_KEY"),
		LastfmUsername:   os.Getenv("LASTFM_USERNAME"),
		MusicBrainzAgent: envOrDefault("MUSICBRAINZ_USER_AGENT", "musiclib/1.0 (personal music library)"),
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
