import { useState, useEffect, useCallback } from 'react';
import { api, Artist, SortOption } from '../lib/api';
import { formatNumber } from '../lib/format';
import { Link } from 'react-router-dom';
import SortBar from '../components/SortBar';

export default function ArtistsPage() {
  const [artists, setArtists] = useState<Artist[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [sort, setSort] = useState<SortOption>('name_asc');
  const [loading, setLoading] = useState(true);
  const limit = 50;

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.listArtists(page, limit, sort);
      setArtists(data.items || []);
      setTotal(data.total);
    } catch (e) {
      console.error('Failed to load artists', e);
    } finally {
      setLoading(false);
    }
  }, [page, sort]);

  useEffect(() => { load(); }, [load]);

  const totalPages = Math.ceil(total / limit);

  const handleSort = (s: SortOption) => {
    setSort(s);
    setPage(1);
  };

  return (
    <div>
      <div className="page-header">
        <h1>Artists</h1>
        <span className="count">{formatNumber(total)}</span>
      </div>

      <SortBar value={sort} onChange={handleSort} />

      {loading ? (
        <div className="loading">Loading...</div>
      ) : (
        <>
          <div className="grid">
            {artists.map(a => (
              <Link key={a.id} to={`/artists/${a.id}`} className="card">
                <div className="card-title">
                  {a.favorite && <span className="fav">&#9733;</span>}
                  {a.name}
                </div>
                <div className="card-meta">
                  {formatNumber(a.scrobble_count)} scrobbles
                  {a.album_count > 0 && <> &middot; {a.album_count} albums</>}
                  {a.track_count > 0 && <> &middot; {a.track_count} tracks</>}
                </div>
              </Link>
            ))}
          </div>

          {totalPages > 1 && (
            <div className="pagination">
              <button disabled={page <= 1} onClick={() => setPage(p => p - 1)}>Prev</button>
              <span>Page {page} of {totalPages}</span>
              <button disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}>Next</button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
