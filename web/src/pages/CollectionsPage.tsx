import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { api, CollectionTree, CollectionSummary } from '../lib/api';

export default function CollectionsPage() {
  const [tree, setTree] = useState<CollectionTree[]>([]);
  const [collections, setCollections] = useState<CollectionSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [showNew, setShowNew] = useState(false);
  const [newName, setNewName] = useState('');
  const [newDesc, setNewDesc] = useState('');

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [treeData, colData] = await Promise.all([
        api.getCollectionTree(),
        api.listCollections(),
      ]);
      setTree(treeData.items || []);
      setCollections(colData.items || []);
    } catch (e) {
      console.error('Failed to load collections', e);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  const createCollection = async () => {
    if (!newName.trim()) return;
    await api.createCollection({ name: newName.trim(), description: newDesc.trim() });
    setNewName('');
    setNewDesc('');
    setShowNew(false);
    load();
  };

  if (loading) return <div className="loading">Loading...</div>;

  const totalCollections = collections.length;

  return (
    <div>
      <div className="page-header">
        <div>
          <h1>Curation</h1>
          <div className="meta">Your personal canon</div>
        </div>
        <button className="btn" onClick={() => setShowNew(true)}>+ New Collection</button>
      </div>

      {showNew && (
        <div className="new-collection-form">
          <input
            type="text"
            placeholder="Collection name"
            value={newName}
            onChange={e => setNewName(e.target.value)}
            autoFocus
            onKeyDown={e => e.key === 'Enter' && createCollection()}
          />
          <input
            type="text"
            placeholder="Description (optional)"
            value={newDesc}
            onChange={e => setNewDesc(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && createCollection()}
          />
          <div className="note-actions">
            <button className="btn" onClick={createCollection}>Create</button>
            <button className="btn btn-secondary" onClick={() => { setShowNew(false); setNewName(''); setNewDesc(''); }}>Cancel</button>
          </div>
        </div>
      )}

      {tree.length === 0 && !showNew ? (
        <div className="empty">
          <p>No collections yet.</p>
          <p style={{ marginTop: '0.5rem' }}>
            Create your first collection to start curating your music library.
          </p>
        </div>
      ) : (
        <div className="collection-tree">
          {tree.map(node => (
            <TreeNode key={node.id} node={node} depth={0} />
          ))}
        </div>
      )}

      {totalCollections > 0 && (
        <div className="meta" style={{ marginTop: '2rem' }}>
          {totalCollections} collection{totalCollections !== 1 ? 's' : ''}
        </div>
      )}
    </div>
  );
}

function TreeNode({ node, depth }: { node: CollectionTree; depth: number }) {
  const [expanded, setExpanded] = useState(depth < 1);
  const hasChildren = node.children && node.children.length > 0;

  return (
    <div className="tree-node">
      <div
        className="tree-row"
        style={{ paddingLeft: `${depth * 1.25}rem` }}
      >
        <button
          className="tree-toggle"
          onClick={() => setExpanded(!expanded)}
          disabled={!hasChildren}
        >
          {hasChildren ? (expanded ? '\u25BC' : '\u25B6') : '\u00B7'}
        </button>
        <Link to={`/collections/${node.id}`} className="tree-link">
          {node.name}
        </Link>
        {hasChildren && (
          <span className="tree-count">{node.children.length}</span>
        )}
      </div>
      {expanded && hasChildren && (
        <div className="tree-children">
          {node.children.map(child => (
            <TreeNode key={child.id} node={child} depth={depth + 1} />
          ))}
        </div>
      )}
    </div>
  );
}
