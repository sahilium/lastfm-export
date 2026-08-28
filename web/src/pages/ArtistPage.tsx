import { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api, Artist, Album, Track, Tag } from '../lib/api';
import { formatNumber, formatDate } from '../lib/format';

export default function ArtistPage() {
  const { id } = useParams<{ id: string }>();
  const [artist, setArtist] = useState<Artist | null>(null);
  const [albums, setAlbums] = useState<Album[]>([]);
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
      const [a, al, tr, tg] = await Promise.all([
        api.getArtist(numId),
        api.getArtistAlbums(numId),
        api.getArtistTracks(numId),
        api.getTags('artist', numId),
      ]);
      setArtist(a);
      setAlbums(al.items || []);
      setTracks(tr.items || []);
      setTags(tg.items || []);
      setNote(a.note || '');
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => { load(); }, [load]);

  const saveNote = async () => {
    if (!id) return;
    await api.updateArtist(parseInt(id), { note });
    setEditing(false);
    load();
  };

  const toggleFavorite = async () => {
    if (!id || !artist) return;
    await api.updateArtist(parseInt(id), { favorite: !artist.favorite });
    load();
  };

  if (loading) return <div className="loading">Loading...</div>;
  if (!artist) return <div className="error">Artist not found</div>;

  return (
    <div>
      <div className="page-header">
        <div>
          <h1>
            {artist.favorite && <span className="fav">&#9733;</span>}
            {artist.name}
          </h1>
          {artist.mbid && (
            <div className="meta">MBID: {artist.mbid}</div>
          )}
        </div>
        <button className="btn" onClick={toggleFavorite}>
          {artist.favorite ? 'Unfavorite' : 'Favorite'}
        </button>
      </div>

      <div className="stats-row">
        <div className="stat">
          <div className="stat-value">{formatNumber(artist.scrobble_count)}</div>
          <div className="stat-label">Scrobbles</div>
        </div>
        <div className="stat">
          <div className="stat-value">{artist.album_count}</div>
          <div className="stat-label">Albums</div>
        </div>
        <div className="stat">
          <div className="stat-value">{artist.track_count}</div>
          <div className="stat-label">Tracks</div>
        </div>
        {artist.first_listened && (
          <div className="stat">
            <div className="stat-value">{formatDate(artist.first_listened)}</div>
            <div className="stat-label">First listened</div>
          </div>
        )}
        {artist.last_listened && (
          <div className="stat">
            <div className="stat-value">{formatDate(artist.last_listened)}</div>
            <div className="stat-label">Last listened</div>
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
              <button className="btn btn-secondary" onClick={() => { setEditing(false); setNote(artist.note || ''); }}>Cancel</button>
            </div>
          </div>
        ) : (
          <div className="note" onClick={() => setEditing(true)}>
            {artist.note || <span className="placeholder">Click to add a note...</span>}
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

      {albums.length > 0 && (
        <section className="section">
          <h2>Albums ({albums.length})</h2>
          <div className="grid">
            {albums.map(al => (
              <Link key={al.id} to={`/albums/${al.id}`} className="card">
                <div className="card-title">
                  {al.favorite && <span className="fav">&#9733;</span>}
                  {al.name}
                </div>
                <div className="card-meta">
                  {formatNumber(al.scrobble_count)} scrobbles
                  {al.release_date && <> &middot; {al.release_date}</>}
                </div>
              </Link>
            ))}
          </div>
        </section>
      )}

      {tracks.length > 0 && (
        <section className="section">
          <h2>Tracks ({tracks.length})</h2>
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
