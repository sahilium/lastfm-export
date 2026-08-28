import { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api, Album, Track, Tag } from '../lib/api';
import { formatNumber, formatDate } from '../lib/format';

export default function AlbumPage() {
  const { id } = useParams<{ id: string }>();
  const [album, setAlbum] = useState<Album | null>(null);
  const [tracks, setTracks] = useState<Track[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [note, setNote] = useState('');
  const [editing, setEditing] = useState(false);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    if (!id) return;
    const numId = parseInt(id);
    setLoading(true);
    try {
      const [al, tr, tg] = await Promise.all([
        api.getAlbum(numId),
        api.getAlbumTracks(numId),
        api.getTags('album', numId),
      ]);
      setAlbum(al);
      setTracks(tr.items || []);
      setTags(tg.items || []);
      setNote(al.note || '');
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => { load(); }, [load]);

  const saveNote = async () => {
    if (!id) return;
    await api.updateAlbum(parseInt(id), { note });
    setEditing(false);
    load();
  };

  const toggleFavorite = async () => {
    if (!id || !album) return;
    await api.updateAlbum(parseInt(id), { favorite: !album.favorite });
    load();
  };

  if (loading) return <div className="loading">Loading...</div>;
  if (!album) return <div className="error">Album not found</div>;

  return (
    <div>
      <div className="page-header">
        <div>
          <h1>
            {album.favorite && <span className="fav">&#9733;</span>}
            {album.name}
          </h1>
          {album.artist_name && (
            <div className="meta">
              <Link to={`/artists/${album.artist_id}`}>{album.artist_name}</Link>
            </div>
          )}
          {album.mbid && (
            <div className="meta">MBID: {album.mbid}</div>
          )}
        </div>
        <button className="btn" onClick={toggleFavorite}>
          {album.favorite ? 'Unfavorite' : 'Favorite'}
        </button>
      </div>

      <div className="stats-row">
        <div className="stat">
          <div className="stat-value">{formatNumber(album.scrobble_count)}</div>
          <div className="stat-label">Scrobbles</div>
        </div>
        <div className="stat">
          <div className="stat-value">{album.track_count}</div>
          <div className="stat-label">Tracks</div>
        </div>
        {album.release_date && (
          <div className="stat">
            <div className="stat-value">{album.release_date}</div>
            <div className="stat-label">Release date</div>
          </div>
        )}
        {album.first_listened && (
          <div className="stat">
            <div className="stat-value">{formatDate(album.first_listened)}</div>
            <div className="stat-label">First listened</div>
          </div>
        )}
      </div>

      <section className="section">
        <h2>Note</h2>
        {editing ? (
          <div className="note-edit">
            <textarea
              value={note}
              onChange={e => setNote(e.target.value)}
              rows={4}
            />
            <div className="note-actions">
              <button className="btn" onClick={saveNote}>Save</button>
              <button className="btn btn-secondary" onClick={() => { setEditing(false); setNote(album.note || ''); }}>Cancel</button>
            </div>
          </div>
        ) : (
          <div className="note" onClick={() => setEditing(true)}>
            {album.note || <span className="placeholder">Click to add a note...</span>}
          </div>
        )}
      </section>

      {tags.length > 0 && (
        <section className="section">
          <h2>Tags</h2>
          <div className="tags">
            {tags.map(t => (
              <span key={t.id} className="tag">{t.name}</span>
            ))}
          </div>
        </section>
      )}

      {tracks.length > 0 && (
        <section className="section">
          <h2>Tracks</h2>
          <div className="list">
            {tracks.map(tr => (
              <Link key={tr.id} to={`/tracks/${tr.id}`} className="list-item">
                <div className="list-item-title">
                  {tr.favorite && <span className="fav">&#9733;</span>}
                  {tr.name}
                </div>
                <div className="list-item-meta">
                  {formatNumber(tr.scrobble_count)} scrobbles
                </div>
              </Link>
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
