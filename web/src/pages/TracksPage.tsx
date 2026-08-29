import { useState, useEffect, useCallback } from 'react';
import { api, Track, SortOption } from '../lib/api';
import { formatNumber } from '../lib/format';
import { Link } from 'react-router-dom';
import SortBar from '../components/SortBar';

export default function TracksPage() {
  const [tracks, setTracks] = useState<Track[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [sort, setSort] = useState<SortOption>('name_asc');
  const [loading, setLoading] = useState(true);
  const limit = 50;

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.listTracks(page, limit, sort);
      setTracks(data.items || []);
      setTotal(data.total);
    } catch (e) {
      console.error('Failed to load tracks', e);
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
        <h1>Tracks</h1>
        <span className="count">{formatNumber(total)}</span>
      </div>

      <SortBar value={sort} onChange={handleSort} />

      {loading ? (
        <div className="loading">Loading...</div>
      ) : (
        <>
          <div className="list">
            {tracks.map(tr => (
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
