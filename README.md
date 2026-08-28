# musiclib

A local-first personal music library web application.

Browse, search, and annotate your music collection. Powered by Last.fm listening history and MusicBrainz metadata.

## Quick start

```bash
# Import existing scrobbles from an SQLite export
./musiclib import db.sqlite3

# Start the server
./musiclib serve
```

Open http://localhost:8080 in your browser.

## CLI commands

```bash
musiclib serve              # Start the web server
musiclib import <source>    # Import scrobbles from an SQLite database
musiclib migrate            # Run database migrations
musiclib stats              # Show library statistics
musiclib fts                # Rebuild full-text search indexes
```

## Configuration

Set via environment variables:

| Variable | Description | Default |
|---|---|---|
| `DATABASE_PATH` | Path to SQLite database | `musiclib.db` |
| `SERVER_ADDRESS` | Server listen address | `:8080` |
| `LASTFM_API_KEY` | Last.fm API key | - |
| `LASTFM_USERNAME` | Last.fm username | - |
| `MUSICBRAINZ_USER_AGENT` | MusicBrainz API user agent | `musiclib/1.0` |
| `TURSO_DATABASE_URL` | Turso database URL (for remote DB) | - |
| `TURSO_AUTH_TOKEN` | Turso auth token | - |

## Development

### Backend

```bash
go run ./cmd/musiclib
go test ./...
go build -o musiclib ./cmd/musiclib
```

### Frontend

```bash
cd web
bun install
bun run dev      # Dev server with hot reload (proxies API to :8080)
bun run build    # Build for production
```

### Production build

```bash
cd web && bun run build && cd ..
go build -o musiclib ./cmd/musiclib
```

The final binary embeds the frontend. No Node/Bun required to run it.

## Architecture

```
HTTP / Gin
    ↓
Handlers
    ↓
Application Services
    ↓
Repositories
    ↓
SQLite (local) or Turso (remote)
```

### Tech stack

- **Backend:** Go, Gin, SQLite (pure-Go via modernc.org/sqlite)
- **Frontend:** React, TypeScript, Vite, Bun
- **Database:** SQLite with FTS5 for search; optional Turso/libSQL for remote
- **Distribution:** Single binary with embedded frontend
