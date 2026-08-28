#!/usr/bin/env python3

import argparse
import os
import sqlite3
import sys
import time
from typing import Optional

import requests


API_URL = "https://ws.audioscrobbler.com/2.0/"
PAGE_SIZE = 200

# Last.fm doesn't publish a numeric rate limit in the API docs.
# Be deliberately conservative.
REQUEST_DELAY = 1.0

MAX_RETRIES = 8
INITIAL_BACKOFF = 2.0


CREATE_TABLE_SQL = """
CREATE TABLE IF NOT EXISTS scrobbles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    artist TEXT NOT NULL,
    artist_mbid TEXT,

    album TEXT,
    album_mbid TEXT,

    track TEXT NOT NULL,
    track_mbid TEXT,

    timestamp INTEGER NOT NULL,

    loved INTEGER,
    url TEXT,

    UNIQUE(
        artist,
        track,
        timestamp
    )
);
"""

CREATE_INDEXES_SQL = [
    """
    CREATE INDEX IF NOT EXISTS idx_scrobbles_timestamp
    ON scrobbles(timestamp);
    """,
    """
    CREATE INDEX IF NOT EXISTS idx_scrobbles_artist
    ON scrobbles(artist);
    """,
    """
    CREATE INDEX IF NOT EXISTS idx_scrobbles_album
    ON scrobbles(album);
    """,
]


def create_db(conn: sqlite3.Connection):
    conn.execute(CREATE_TABLE_SQL)

    for sql in CREATE_INDEXES_SQL:
        conn.execute(sql)

    conn.commit()


def request_page(
    session: requests.Session,
    username: str,
    api_key: str,
    page: int,
) -> dict:

    params = {
        "method": "user.getRecentTracks",
        "user": username,
        "api_key": api_key,
        "format": "json",
        "limit": PAGE_SIZE,
        "page": page,

        # Extended gives us the "loved" field.
        "extended": 1,
    }

    backoff = INITIAL_BACKOFF

    for attempt in range(MAX_RETRIES):

        try:
            response = session.get(
                API_URL,
                params=params,
                timeout=30,
            )

            response.raise_for_status()

            data = response.json()

            # Last.fm sometimes returns HTTP 200 with an API error.
            if "error" in data:
                error_code = data["error"]

                if str(error_code) == "29":
                    print(
                        f"Rate limited. Sleeping {backoff:.1f}s...",
                        file=sys.stderr,
                    )
                    time.sleep(backoff)
                    backoff *= 2
                    continue

                raise RuntimeError(
                    f"Last.fm API error {error_code}: "
                    f"{data.get('message', 'unknown error')}"
                )

            return data

        except (
            requests.RequestException,
            ValueError,
        ) as exc:

            if attempt == MAX_RETRIES - 1:
                raise

            print(
                f"Request failed: {exc}. "
                f"Retrying in {backoff:.1f}s...",
                file=sys.stderr,
            )

            time.sleep(backoff)
            backoff *= 2

    raise RuntimeError("Failed to fetch page")


def parse_track(track: dict) -> Optional[tuple]:
    """
    Convert a Last.fm track object into a DB row.

    Now-playing tracks don't have a timestamp, so ignore them.
    """

    date = track.get("date")

    if not date:
        return None

    timestamp = date.get("uts")

    if not timestamp:
        return None

    artist = track.get("artist", {})

    # Depending on the response shape, loved may be on the track.
    loved = track.get("loved")

    if loved is not None:
        loved = int(loved)

    return (
        artist.get("#text") or artist.get("name") or "",
        artist.get("mbid") or None,

        track.get("album", {}).get("#text") or None,
        track.get("album", {}).get("mbid") or None,

        track.get("name") or "",
        track.get("mbid") or None,

        int(timestamp),

        loved,

        track.get("url") or None,
    )


def insert_tracks(
    conn: sqlite3.Connection,
    tracks: list[tuple],
) -> int:

    sql = """
    INSERT OR IGNORE INTO scrobbles (
        artist,
        artist_mbid,
        album,
        album_mbid,
        track,
        track_mbid,
        timestamp,
        loved,
        url
    )
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    """

    before = conn.total_changes

    conn.executemany(sql, tracks)
    conn.commit()

    return conn.total_changes - before


def main():
    parser = argparse.ArgumentParser(
        description="Export a Last.fm scrobble history into SQLite."
    )

    parser.add_argument(
        "--username",
        required=True,
        help="Your Last.fm username",
    )

    parser.add_argument(
        "--api-key",
        default=os.environ.get("LASTFM_API_KEY"),
        help="Last.fm API key. Can also use LASTFM_API_KEY.",
    )

    parser.add_argument(
        "--db",
        default="lastfm.sqlite3",
        help="SQLite database filename",
    )

    args = parser.parse_args()

    if not args.api_key:
        parser.error(
            "Provide --api-key or set LASTFM_API_KEY"
        )

    conn = sqlite3.connect(args.db)

    create_db(conn)

    session = requests.Session()

    # First request tells us how many pages exist.
    print("Fetching Last.fm history...")

    first_page = request_page(
        session,
        args.username,
        args.api_key,
        page=1,
    )

    recenttracks = first_page["recenttracks"]

    attr = recenttracks.get("@attr", {})

    total_pages = int(attr.get("totalPages", 1))
    total_tracks = int(attr.get("total", 0))

    print(f"Last.fm reports {total_tracks:,} scrobbles.")
    print(f"Pages: {total_pages:,}")
    print()

    inserted = 0
    processed = 0

    # Process page 1.
    tracks = []

    for track in recenttracks.get("track", []):
        parsed = parse_track(track)

        if parsed:
            tracks.append(parsed)

    inserted += insert_tracks(conn, tracks)
    processed += len(tracks)

    print(
        f"[1/{total_pages}] "
        f"processed={processed:,}, "
        f"new={inserted:,}"
    )

    # Remaining pages.
    for page in range(2, total_pages + 1):

        time.sleep(REQUEST_DELAY)

        data = request_page(
            session,
            args.username,
            args.api_key,
            page,
        )

        tracks = []

        for track in data["recenttracks"].get("track", []):
            parsed = parse_track(track)

            if parsed:
                tracks.append(parsed)

        inserted_now = insert_tracks(conn, tracks)

        inserted += inserted_now
        processed += len(tracks)

        print(
            f"[{page}/{total_pages}] "
            f"processed={processed:,}, "
            f"new={inserted_now:,}, "
            f"total_new={inserted:,}"
        )

    conn.close()

    print()
    print("Done.")
    print(f"Database: {args.db}")
    print(f"Processed: {processed:,}")
    print(f"New rows: {inserted:,}")


if __name__ == "__main__":
    main()