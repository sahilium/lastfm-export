import { useState, useEffect, useRef, useCallback } from 'react';
import { api, CollectionSummary, Artist, Album, Track } from '../lib/api';

interface Props {
  onSelect: (itemType: string, itemId: number) => void;
  onClose: () => void;
}

type ResultItem = {
  type: string;
  id: number;
  name: string;
  meta: string;
};

type ResultGroup = {
  type: string;
  label: string;
  items: ResultItem[];
};

export default function AddToCollectionPicker({ onSelect, onClose }: Props) {
  const [query, setQuery] = useState('');
  const [groups, setGroups] = useState<ResultGroup[]>([]);
  const [collections, setCollections] = useState<CollectionSummary[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedType, setSelectedType] = useState<string | null>(null);
  const [activeIdx, setActiveIdx] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const allItems = groups.flatMap(g => g.items);

  useEffect(() => {
    inputRef.current?.focus();
    api.listCollections().then(d => setCollections(d.items || [])).catch(() => {});
  }, []);

  const search = useCallback(async (q: string) => {
    if (!q.trim()) {
      setGroups([]);
      setActiveIdx(0);
      return;
    }
    setLoading(true);
    try {
      const data = await api.search(q);
      const resultGroups: ResultGroup[] = [];

      if (selectedType === null || selectedType === 'artist') {
        const items: ResultItem[] = (data.artists?.items || []).map((a: Artist) => ({
          type: 'artist', id: a.id, name: a.name, meta: `${a.scrobble_count} scrobbles`,
        }));
        if (items.length > 0) resultGroups.push({ type: 'artist', label: 'Artists', items });
      }

      if (selectedType === null || selectedType === 'album') {
        const items: ResultItem[] = (data.albums?.items || []).map((a: Album) => ({
          type: 'album', id: a.id, name: a.name, meta: a.artist_name || 'Album',
        }));
        if (items.length > 0) resultGroups.push({ type: 'album', label: 'Albums', items });
      }

      if (selectedType === null || selectedType === 'track') {
        const items: ResultItem[] = (data.tracks?.items || []).map((t: Track) => ({
          type: 'track', id: t.id, name: t.name, meta: t.artist_name || 'Track',
        }));
        if (items.length > 0) resultGroups.push({ type: 'track', label: 'Tracks', items });
      }

      setGroups(resultGroups);
      setActiveIdx(0);
    } catch (e) {
      console.error('Search failed', e);
    } finally {
      setLoading(false);
    }
  }, [selectedType]);

  useEffect(() => {
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => search(query), 200);
    return () => { if (timerRef.current) clearTimeout(timerRef.current); };
  }, [query, search]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setActiveIdx(i => Math.min(i + 1, allItems.length - 1));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setActiveIdx(i => Math.max(i - 1, 0));
    } else if (e.key === 'Enter' && allItems[activeIdx]) {
      e.preventDefault();
      const item = allItems[activeIdx];
      onSelect(item.type, item.id);
    } else if (e.key === 'Escape') {
      onClose();
    }
  };

  const showCollections = !query && collections.length > 0;
  const totalResults = allItems.length;

  return (
    <div className="picker-overlay" onClick={onClose}>
      <div className="picker" onClick={e => e.stopPropagation()} onKeyDown={handleKeyDown}>
        <div className="picker-header">
          <h3>Add to collection</h3>
          <button className="picker-close" onClick={onClose}>&times;</button>
        </div>

        <div className="picker-toolbar">
          <div className="search-input" style={{ margin: 0, flex: 1 }}>
            <input
              ref={inputRef}
              type="text"
              placeholder="Search artists, albums, tracks..."
              value={query}
              onChange={e => setQuery(e.target.value)}
            />
          </div>
          <div className="picker-filters">
            {(['artist', 'album', 'track'] as const).map(t => (
              <button
                key={t}
                className={`filter-btn ${selectedType === t ? 'active' : ''}`}
                onClick={() => setSelectedType(selectedType === t ? null : t)}
              >
                {t}s
              </button>
            ))}
          </div>
        </div>

        <div className="picker-body">
          {showCollections && (
            <div className="picker-section">
              <div className="picker-section-label">Your collections</div>
              {collections.map(c => (
                <button
                  key={c.id}
                  className="picker-item"
                  onClick={() => onSelect('collection', c.id)}
                >
                  <span className="picker-item-name">{c.name}</span>
                  <span className="picker-item-meta type-collection">{c.item_count} items</span>
                </button>
              ))}
            </div>
          )}

          {groups.map(group => (
            <div key={group.type} className="picker-section">
              <div className="picker-section-label">
                {group.label}
                <span className="picker-section-count">{group.items.length}</span>
              </div>
              {group.items.map(item => {
                const idx = allItems.indexOf(item);
                return (
                  <button
                    key={`${item.type}-${item.id}`}
                    className={`picker-item ${idx === activeIdx ? 'active' : ''}`}
                    onClick={() => onSelect(item.type, item.id)}
                    onMouseEnter={() => setActiveIdx(idx)}
                  >
                    <span className={`picker-item-badge type-${item.type}`}>
                      {item.type[0].toUpperCase()}
                    </span>
                    <div className="picker-item-info">
                      <span className="picker-item-name">{item.name}</span>
                      <span className="picker-item-meta">{item.meta}</span>
                    </div>
                  </button>
                );
              })}
            </div>
          ))}

          {query && totalResults === 0 && !loading && (
            <div className="picker-empty">No results for "{query}"</div>
          )}

          {loading && <div className="picker-loading">Searching...</div>}

          {!query && !showCollections && (
            <div className="picker-empty">Type to search your library</div>
          )}
        </div>

        {totalResults > 0 && (
          <div className="picker-footer">
            <span className="picker-hint">{totalResults} result{totalResults !== 1 ? 's' : ''}</span>
            <span className="picker-hint">&uarr;&darr; navigate &middot; &crarr; select</span>
          </div>
        )}
      </div>
    </div>
  );
}
