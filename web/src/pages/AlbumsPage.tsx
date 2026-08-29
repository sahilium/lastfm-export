import { useState, useEffect, useCallback } from 'react';
import { api, Album, SortOption } from '../lib/api';
import { formatNumber } from '../lib/format';
import { Link } from 'react-router-dom';
import SortBar from '../components/SortBar';

export default function AlbumsPage() {
  const [albums, setAlbums] = useState<Album[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [sort, setSort] = useState<SortOption>('name_asc');
  const [loading, setLoading] = useState(true);
  const limit = 50;

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await api.listAlbums(page, limit, sort);
      setAlbums(data.items || []);
      setTotal(data.total);
    } catch (e) {
      console.error('Failed to load albums', e);
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
        <h1>Albums</h1>
        <span className="count">{formatNumber(total)}</span>
      </div>

      <SortBar value={sort} onChange={handleSort} />

      {loading ? (
        <div className="loading">Loading...</div>
      ) : (
        <>
          <div className="grid">
            {albums.map(al => (
              <Link key={al.id} to={`/albums/${al.id}`} className="card">
                <div className="card-title">
                  {al.favorite && <span className="fav">&#9733;</span>}
                  {al.name}
                </div>
                <div className="card-meta">
                  {al.artist_name && <>{al.artist_name} &middot;</>}
                  {formatNumber(al.scrobble_count)} scrobbles
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
