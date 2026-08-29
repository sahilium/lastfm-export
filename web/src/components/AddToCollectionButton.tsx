import { useState } from 'react';
import { api, CollectionSummary } from '../lib/api';

interface Props {
  entityType: string;
  entityId: number;
  entityName: string;
}

export default function AddToCollectionButton({ entityType, entityId, entityName }: Props) {
  const [showPicker, setShowPicker] = useState(false);
  const [collections, setCollections] = useState<CollectionSummary[]>([]);
  const [loading, setLoading] = useState(false);
  const [added, setAdded] = useState<string | null>(null);

  const open = async () => {
    setLoading(true);
    try {
      const data = await api.listCollections();
      setCollections(data.items || []);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
      setShowPicker(true);
    }
  };

  const addTo = async (collectionId: number, collectionName: string) => {
    try {
      await api.addCollectionItem(collectionId, { item_type: entityType, item_id: entityId });
      setAdded(collectionName);
      setTimeout(() => setAdded(null), 2000);
      setShowPicker(false);
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <>
      <button className="btn btn-secondary" onClick={open}>
        + Add to collection
      </button>
      {added && (
        <span className="added-toast">Added to {added}</span>
      )}
      {showPicker && (
        <div className="picker-overlay" onClick={() => setShowPicker(false)}>
          <div className="picker" onClick={e => e.stopPropagation()}>
            <div className="picker-header">
              <h3>Add {entityName} to...</h3>
              <button className="picker-close" onClick={() => setShowPicker(false)}>&times;</button>
            </div>
            {loading ? (
              <div className="picker-loading">Loading collections...</div>
            ) : collections.length === 0 ? (
              <div className="picker-empty">No collections yet. Create one first.</div>
            ) : (
              <div className="picker-section">
                {collections.map(c => (
                  <button
                    key={c.id}
                    className="picker-item"
                    onClick={() => addTo(c.id, c.name)}
                  >
                    <span className="picker-item-name">{c.name}</span>
                    <span className="picker-item-type type-collection">
                      {c.item_count} items
                    </span>
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
      )}
    </>
  );
}
