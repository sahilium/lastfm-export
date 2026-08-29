import { useState, useEffect, useCallback, useRef } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api, Collection, CollectionItem } from '../lib/api';
import AddToCollectionPicker from '../components/AddToCollectionPicker';

export default function CollectionPage() {
  const { id } = useParams<{ id: string }>();
  const [collection, setCollection] = useState<Collection | null>(null);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [editName, setEditName] = useState('');
  const [editDesc, setEditDesc] = useState('');
  const [showAdd, setShowAdd] = useState(false);
  const [showNewChild, setShowNewChild] = useState(false);
  const [childName, setChildName] = useState('');
  const [childDesc, setChildDesc] = useState('');
  const [dragIdx, setDragIdx] = useState<number | null>(null);
  const [editingNote, setEditingNote] = useState<number | null>(null);
  const [noteText, setNoteText] = useState('');
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [movingItem, setMovingItem] = useState<number | null>(null);

  const load = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    try {
      const data = await api.getCollection(parseInt(id));
      setCollection(data);
    } catch (e) {
      console.error('Failed to load collection', e);
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => { load(); }, [load]);

  const saveEdit = async () => {
    if (!id || !editName.trim()) return;
    await api.updateCollection(parseInt(id), { name: editName.trim(), description: editDesc.trim() });
    setEditing(false);
    load();
  };

  const deleteCollection = async () => {
    if (!id) return;
    await api.deleteCollection(parseInt(id));
    window.location.href = '/collections';
  };

  const addChild = async () => {
    if (!id || !childName.trim()) return;
    await api.createCollection({ name: childName.trim(), description: childDesc.trim(), parent_id: parseInt(id) });
    setChildName('');
    setChildDesc('');
    setShowNewChild(false);
    load();
  };

  const addItem = async (itemType: string, itemId: number) => {
    if (!id) return;
    await api.addCollectionItem(parseInt(id), { item_type: itemType, item_id: itemId });
    setShowAdd(false);
    load();
  };

  const removeItem = async (itemId: number) => {
    if (!id) return;
    await api.removeCollectionItem(parseInt(id), itemId);
    load();
  };

  const saveItemNote = async (itemId: number) => {
    if (!id) return;
    await api.updateCollectionItemNote(parseInt(id), itemId, noteText);
    setEditingNote(null);
    load();
  };

  const handleDragStart = (idx: number) => setDragIdx(idx);

  const handleDragOver = (e: React.DragEvent, idx: number) => {
    e.preventDefault();
    if (dragIdx === null || dragIdx === idx || !collection) return;
    const items = [...collection.items];
    const [moved] = items.splice(dragIdx, 1);
    items.splice(idx, 0, moved);
    setCollection({ ...collection, items });
  };

  const handleDragEnd = async () => {
    setDragIdx(null);
    if (!id || !collection) return;
    const positions = collection.items.map((item, idx) => ({ id: item.id, position: idx }));
    await api.reorderCollectionItems(parseInt(id), positions);
  };

  if (loading) return <div className="loading">Loading...</div>;
  if (!collection) return <div className="error">Collection not found</div>;

  const breadcrumbs: { id: number; name: string }[] = [];
  // Build breadcrumbs from collection name only (parent chain not in response)
  breadcrumbs.push({ id: collection.id, name: collection.name });

  return (
    <div>
      <div className="page-header">
        <div>
          {editing ? (
            <div className="collection-edit">
              <input
                type="text"
                value={editName}
                onChange={e => setEditName(e.target.value)}
                autoFocus
                onKeyDown={e => e.key === 'Enter' && saveEdit()}
                className="collection-edit-name"
              />
              <input
                type="text"
                value={editDesc}
                onChange={e => setEditDesc(e.target.value)}
                placeholder="Description (optional)"
                onKeyDown={e => e.key === 'Enter' && saveEdit()}
                className="collection-edit-desc"
              />
              <div className="note-actions">
                <button className="btn" onClick={saveEdit}>Save</button>
                <button className="btn btn-secondary" onClick={() => setEditing(false)}>Cancel</button>
              </div>
            </div>
          ) : (
            <>
              <h1>{collection.name}</h1>
              {collection.description && (
                <div className="meta">{collection.description}</div>
              )}
            </>
          )}
        </div>
        <div className="page-header-actions">
          {!editing && (
            <button
              className="btn btn-secondary"
              onClick={() => {
                setEditing(true);
                setEditName(collection.name);
                setEditDesc(collection.description);
              }}
            >
              Edit
            </button>
          )}
          {!confirmDelete ? (
            <button className="btn btn-secondary" onClick={() => setConfirmDelete(true)}>
              Delete
            </button>
          ) : (
            <div className="confirm-delete">
              <span>Delete this collection?</span>
              <button className="btn" onClick={deleteCollection}>Yes</button>
              <button className="btn btn-secondary" onClick={() => setConfirmDelete(false)}>No</button>
            </div>
          )}
        </div>
      </div>

      <div className="collection-actions">
        <button className="btn" onClick={() => setShowAdd(true)}>+ Add</button>
        <button className="btn btn-secondary" onClick={() => setShowNewChild(true)}>+ Subcollection</button>
      </div>

      {showNewChild && (
        <div className="new-collection-form">
          <input
            type="text"
            placeholder="Subcollection name"
            value={childName}
            onChange={e => setChildName(e.target.value)}
            autoFocus
            onKeyDown={e => e.key === 'Enter' && addChild()}
          />
          <input
            type="text"
            placeholder="Description (optional)"
            value={childDesc}
            onChange={e => setChildDesc(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && addChild()}
          />
          <div className="note-actions">
            <button className="btn" onClick={addChild}>Create</button>
            <button className="btn btn-secondary" onClick={() => { setShowNewChild(false); setChildName(''); setChildDesc(''); }}>Cancel</button>
          </div>
        </div>
      )}

      {showAdd && (
        <AddToCollectionPicker
          onSelect={addItem}
          onClose={() => setShowAdd(false)}
        />
      )}

      {collection.items.length === 0 ? (
        <div className="empty">
          <p>This collection is empty.</p>
          <p style={{ marginTop: '0.5rem' }}>
            Click + Add to add artists, albums, tracks, or subcollections.
          </p>
        </div>
      ) : (
        <div className="collection-items">
          {collection.items.map((item, idx) => (
            <div
              key={item.id}
              className={`collection-item ${dragIdx === idx ? 'dragging' : ''}`}
              draggable
              onDragStart={() => handleDragStart(idx)}
              onDragOver={(e) => handleDragOver(e, idx)}
              onDragEnd={handleDragEnd}
            >
              <div className="collection-item-grip" title="Drag to reorder">&#9776;</div>
              <div className="collection-item-number">{idx + 1}</div>
              <div className="collection-item-content">
                <div className="collection-item-header">
                  <Link to={getItemLink(item)} className="collection-item-name">
                    {item.name}
                  </Link>
                  <span className={`collection-item-type type-${item.item_type}`}>
                    {item.item_type}
                  </span>
                </div>
                {item.note && editingNote !== item.id && (
                  <div className="collection-item-note" onClick={() => { setEditingNote(item.id); setNoteText(item.note); }}>
                    {item.note}
                  </div>
                )}
                {editingNote === item.id && (
                  <div className="collection-item-note-edit">
                    <textarea
                      value={noteText}
                      onChange={e => setNoteText(e.target.value)}
                      rows={2}
                      autoFocus
                    />
                    <div className="note-actions">
                      <button className="btn" onClick={() => saveItemNote(item.id)}>Save</button>
                      <button className="btn btn-secondary" onClick={() => setEditingNote(null)}>Cancel</button>
                    </div>
                  </div>
                )}
                {!item.note && editingNote !== item.id && (
                  <button
                    className="collection-item-add-note"
                    onClick={() => { setEditingNote(item.id); setNoteText(''); }}
                  >
                    + Add note
                  </button>
                )}
              </div>
              <button
                className="collection-item-remove"
                onClick={() => removeItem(item.id)}
                title="Remove from collection"
              >
                &times;
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function getItemLink(item: CollectionItem): string {
  switch (item.item_type) {
    case 'artist': return `/artists/${item.item_id}`;
    case 'album': return `/albums/${item.item_id}`;
    case 'track': return `/tracks/${item.item_id}`;
    case 'collection': return `/collections/${item.item_id}`;
    default: return '#';
  }
}
