import { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api, Track, Scrobble, Tag } from '../lib/api';
import { formatNumber, formatDateTime } from '../lib/format';

export default function TrackPage() {
  const { id } = useParams<{ id: string }>();
  const [track, setTrack] = useState<Track | null>(null);
  const [history, setHistory] = useState<Scrobble[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [note, setNote] = useState('');
  const [editing, setEditing] = useState(false);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    if (!id) return;
    const numId = parseInt(id);
    setLoading(true);
    try {
      const [tr, hist, tg] = await Promise.all([
        api.getTrack(numId),
        api.getTrackHistory(numId),
        api.getTags('track', numId),
      ]);
      setTrack(tr);
      setHistory(hist.items || []);
      setTags(tg.items || []);
      setNote(tr.note || '');
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => { load(); }, [load]);

  const saveNote = async () => {
    if (!id) return;
    await api.updateTrack(parseInt(id), { note });
    setEditing(false);
    load();
  };

  const toggleFavorite = async () => {
    if (!id || !track) return;
    await api.updateTrack(parseInt(id), { favorite: !track.favorite });
    load();
  };

  if (loading) return <div className="loading">Loading...</div>;
  if (!track) return <div className="error">Track not found</div>;

  return (
    <div>
      <div className="page-header">
        <div>
          <h1>
            {track.favorite && <span className="fav">&#9733;</span>}
            {track.name}
          </h1>
          <div className="meta">
            {track.artist_name && (
              <Link to={`/artists/${track.artist_id}`}>{track.artist_name}</Link>
            )}
            {track.album_name && track.album_id && (
              <> &middot; <Link to={`/albums/${track.album_id}`}>{track.album_name}</Link></>
            )}
          </div>
          {track.mbid && (
            <div className="meta">MBID: {track.mbid}</div>
          )}
        </div>
        <button className="btn" onClick={toggleFavorite}>
          {track.favorite ? 'Unfavorite' : 'Favorite'}
        </button>
      </div>

      <div className="stats-row">
        <div className="stat">
          <div className="stat-value">{formatNumber(track.scrobble_count)}</div>
          <div className="stat-label">Scrobbles</div>
        </div>
        {track.first_listened && (
          <div className="stat">
            <div className="stat-value">{formatDateTime(track.first_listened)}</div>
            <div className="stat-label">First listened</div>
          </div>
        )}
        {track.last_listened && (
          <div className="stat">
            <div className="stat-value">{formatDateTime(track.last_listened)}</div>
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
              <button className="btn btn-secondary" onClick={() => { setEditing(false); setNote(track.note || ''); }}>Cancel</button>
            </div>
          </div>
        ) : (
          <div className="note" onClick={() => setEditing(true)}>
            {track.note || <span className="placeholder">Click to add a note...</span>}
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

      {history.length > 0 && (
        <section className="section">
          <h2>Listening History ({history.length})</h2>
          <div className="list">
            {history.map(s => (
              <div key={s.id} className="list-item">
                <div className="list-item-title">
                  {formatDateTime(s.timestamp)}
                  {s.loved === 1 && <span className="fav"> &#9829;</span>}
                </div>
                <div className="list-item-meta">
                  {s.album && <>{s.album} &middot;</>}
                  {s.url && <a href={s.url} target="_blank" rel="noopener noreferrer">Last.fm</a>}
                </div>
              </div>
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
