import { useState, useEffect, useRef } from 'react';
import { api, SearchResults } from '../lib/api';
import { formatNumber } from '../lib/format';
import { Link } from 'react-router-dom';

export default function SearchPage() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<SearchResults | null>(null);
  const [loading, setLoading] = useState(false);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (!query.trim()) {
      setResults(null);
      return;
    }

    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }

    debounceRef.current = setTimeout(async () => {
      setLoading(true);
      try {
        const data = await api.search(query.trim());
        setResults(data);
      } catch (e) {
        console.error('Search failed', e);
      } finally {
        setLoading(false);
      }
    }, 300);

    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
    };
  }, [query]);

  return (
    <div>
      <div className="page-header">
        <h1>Search</h1>
      </div>

      <div className="search-input">
        <input
          type="text"
          value={query}
          onChange={e => setQuery(e.target.value)}
          placeholder="Search artists, albums, tracks..."
          autoFocus
        />
      </div>

      {loading && <div className="loading">Searching...</div>}

      {results && !loading && (
        <div className="search-results">
          {results.artists.items.length > 0 && (
            <section className="section">
              <h2>Artists ({results.artists.total})</h2>
              <div className="list">
                {results.artists.items.map(a => (
                  <Link key={a.id} to={`/artists/${a.id}`} className="list-item">
                    <div className="list-item-title">
                      {a.favorite && <span className="fav">&#9733;</span>}
                      {a.name}
                    </div>
                    <div className="list-item-meta">
                      {formatNumber(a.scrobble_count)} scrobbles
                    </div>
                  </Link>
                ))}
              </div>
            </section>
          )}

          {results.albums.items.length > 0 && (
            <section className="section">
              <h2>Albums ({results.albums.total})</h2>
              <div className="list">
                {results.albums.items.map(al => (
                  <Link key={al.id} to={`/albums/${al.id}`} className="list-item">
                    <div className="list-item-title">
                      {al.favorite && <span className="fav">&#9733;</span>}
                      {al.name}
                    </div>
                    <div className="list-item-meta">
                      {al.artist_name && <>{al.artist_name} &middot;</>}
                      {formatNumber(al.scrobble_count)} scrobbles
                    </div>
                  </Link>
                ))}
              </div>
            </section>
          )}

          {results.tracks.items.length > 0 && (
            <section className="section">
              <h2>Tracks ({results.tracks.total})</h2>
              <div className="list">
                {results.tracks.items.map(tr => (
                  <Link key={tr.id} to={`/tracks/${tr.id}`} className="list-item">
                    <div className="list-item-title">
                      {tr.favorite && <span className="fav">&#9733;</span>}
                      {tr.name}
                    </div>
                    <div className="list-item-meta">
                      {tr.artist_name && <>{tr.artist_name}</>}
                      {tr.album_name && <> &middot; {tr.album_name}</>}
                      <> &middot; {formatNumber(tr.scrobble_count)} scrobbles</>
                    </div>
                  </Link>
                ))}
              </div>
            </section>
          )}

          {results.artists.items.length === 0 &&
           results.albums.items.length === 0 &&
           results.tracks.items.length === 0 && (
            <div className="empty">No results found</div>
          )}
        </div>
      )}
    </div>
  );
}
