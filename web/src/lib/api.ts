const BASE = '/api';

async function fetchJSON<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`API error: ${res.status}`);
  }
  return res.json();
}

async function patchJSON<T>(url: string, body: unknown): Promise<T> {
  const res = await fetch(url, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    throw new Error(`API error: ${res.status}`);
  }
  return res.json();
}

async function postJSON<T>(url: string, body?: unknown): Promise<T> {
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    throw new Error(`API error: ${res.status}`);
  }
  return res.json();
}

export interface Artist {
  id: number;
  name: string;
  mbid: string | null;
  note: string;
  favorite: boolean;
  created_at: string;
  updated_at: string;
  scrobble_count: number;
  album_count: number;
  track_count: number;
  first_listened: number | null;
  last_listened: number | null;
}

export interface Album {
  id: number;
  artist_id: number;
  name: string;
  mbid: string | null;
  release_date: string | null;
  note: string;
  favorite: boolean;
  created_at: string;
  updated_at: string;
  artist_name: string | null;
  scrobble_count: number;
  track_count: number;
  first_listened: number | null;
  last_listened: number | null;
}

export interface Track {
  id: number;
  artist_id: number;
  album_id: number | null;
  name: string;
  mbid: string | null;
  note: string;
  favorite: boolean;
  created_at: string;
  updated_at: string;
  artist_name: string | null;
  album_name: string | null;
  scrobble_count: number;
  first_listened: number | null;
  last_listened: number | null;
}

export interface Scrobble {
  id: number;
  track_id: number | null;
  artist: string;
  artist_mbid: string | null;
  album: string | null;
  album_mbid: string | null;
  track: string;
  track_mbid: string | null;
  timestamp: number;
  loved: number | null;
  url: string | null;
}

export interface Tag {
  id: number;
  name: string;
}

export interface Paginated<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
}

export interface SearchResults {
  artists: Paginated<Artist>;
  albums: Paginated<Album>;
  tracks: Paginated<Track>;
}

export interface Summary {
  artists: number;
  albums: number;
  tracks: number;
  scrobbles: number;
}

export interface Collection {
  id: number;
  parent_id: number | null;
  name: string;
  description: string;
  created_at: number;
  updated_at: number;
  items: CollectionItem[];
}

export interface CollectionItem {
  id: number;
  item_type: string;
  item_id: number;
  position: number;
  note: string;
  name: string;
}

export interface CollectionSummary {
  id: number;
  parent_id: number | null;
  name: string;
  description: string;
  child_count: number;
  item_count: number;
}

export interface CollectionTree {
  id: number;
  name: string;
  children: CollectionTree[];
}

export type SortOption = 'name_asc' | 'name_desc' | 'scrobbles_asc' | 'scrobbles_desc' | 'recent_asc' | 'recent_desc';

export const api = {
  summary: () => fetchJSON<Summary>(`${BASE}/summary`),

  listArtists: (page = 1, limit = 50, sort: SortOption = 'name_asc') =>
    fetchJSON<Paginated<Artist>>(`${BASE}/artists?page=${page}&limit=${limit}&sort=${sort}`),
  getArtist: (id: number) => fetchJSON<Artist>(`${BASE}/artists/${id}`),
  updateArtist: (id: number, data: { note?: string; favorite?: boolean }) =>
    patchJSON<{ ok: boolean }>(`${BASE}/artists/${id}`, data),
  getArtistAlbums: (id: number) =>
    fetchJSON<{ items: Album[] }>(`${BASE}/artists/${id}/albums`),
  getArtistTracks: (id: number) =>
    fetchJSON<{ items: Track[] }>(`${BASE}/artists/${id}/tracks`),

  listAlbums: (page = 1, limit = 50, sort: SortOption = 'name_asc') =>
    fetchJSON<Paginated<Album>>(`${BASE}/albums?page=${page}&limit=${limit}&sort=${sort}`),
  getAlbum: (id: number) => fetchJSON<Album>(`${BASE}/albums/${id}`),
  updateAlbum: (id: number, data: { note?: string; favorite?: boolean }) =>
    patchJSON<{ ok: boolean }>(`${BASE}/albums/${id}`, data),
  getAlbumTracks: (id: number) =>
    fetchJSON<{ items: Track[] }>(`${BASE}/albums/${id}/tracks`),

  listTracks: (page = 1, limit = 50, sort: SortOption = 'name_asc') =>
    fetchJSON<Paginated<Track>>(`${BASE}/tracks?page=${page}&limit=${limit}&sort=${sort}`),
  getTrack: (id: number) => fetchJSON<Track>(`${BASE}/tracks/${id}`),
  updateTrack: (id: number, data: { note?: string; favorite?: boolean }) =>
    patchJSON<{ ok: boolean }>(`${BASE}/tracks/${id}`, data),
  getTrackHistory: (id: number) =>
    fetchJSON<{ items: Scrobble[] }>(`${BASE}/tracks/${id}/history`),

  search: (q: string) => fetchJSON<SearchResults>(`${BASE}/search?q=${encodeURIComponent(q)}`),

  getTags: (type: string, id: number) =>
    fetchJSON<{ items: Tag[] }>(`${BASE}/tags/${type}/${id}`),
  addTag: (type: string, id: number, name: string) =>
    postJSON<{ ok: boolean }>(`${BASE}/tags/${type}/${id}`, { name }),
  removeTag: (type: string, id: number, tagId: number) =>
    fetch(`${BASE}/tags/${type}/${id}/${tagId}`, { method: 'DELETE' }).then(() => ({ ok: true })),

  syncLastfm: () => postJSON<{ status: string }>(`${BASE}/sync/lastfm`),

  listCollections: () =>
    fetchJSON<{ items: CollectionSummary[] }>(`${BASE}/collections`),
  getCollectionTree: () =>
    fetchJSON<{ items: CollectionTree[] }>(`${BASE}/collections/tree`),
  getCollection: (id: number) =>
    fetchJSON<Collection>(`${BASE}/collections/${id}`),
  createCollection: (data: { name: string; description?: string; parent_id?: number }) =>
    postJSON<Collection>(`${BASE}/collections`, data),
  updateCollection: (id: number, data: { name: string; description?: string }) =>
    patchJSON<{ ok: boolean }>(`${BASE}/collections/${id}`, data),
  deleteCollection: (id: number) =>
    fetch(`${BASE}/collections/${id}`, { method: 'DELETE' }).then(() => ({ ok: true })),
  moveCollection: (id: number, parentId: number | null) =>
    postJSON<{ ok: boolean }>(`${BASE}/collections/${id}/move`, { parent_id: parentId }),
  addCollectionItem: (collectionId: number, data: { item_type: string; item_id: number; note?: string }) =>
    postJSON<CollectionItem>(`${BASE}/collections/${collectionId}/items`, data),
  removeCollectionItem: (collectionId: number, itemId: number) =>
    fetch(`${BASE}/collections/${collectionId}/items/${itemId}`, { method: 'DELETE' }).then(() => ({ ok: true })),
  updateCollectionItemNote: (collectionId: number, itemId: number, note: string) =>
    patchJSON<{ ok: boolean }>(`${BASE}/collections/${collectionId}/items/${itemId}`, { note }),
  reorderCollectionItems: (collectionId: number, items: { id: number; position: number }[]) =>
    postJSON<{ ok: boolean }>(`${BASE}/collections/${collectionId}/items/reorder`, { items }),
};
