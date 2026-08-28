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

export const api = {
  summary: () => fetchJSON<Summary>(`${BASE}/summary`),

  listArtists: (page = 1, limit = 50) =>
    fetchJSON<Paginated<Artist>>(`${BASE}/artists?page=${page}&limit=${limit}`),
  getArtist: (id: number) => fetchJSON<Artist>(`${BASE}/artists/${id}`),
  updateArtist: (id: number, data: { note?: string; favorite?: boolean }) =>
    patchJSON<{ ok: boolean }>(`${BASE}/artists/${id}`, data),
  getArtistAlbums: (id: number) =>
    fetchJSON<{ items: Album[] }>(`${BASE}/artists/${id}/albums`),
  getArtistTracks: (id: number) =>
    fetchJSON<{ items: Track[] }>(`${BASE}/artists/${id}/tracks`),

  listAlbums: (page = 1, limit = 50) =>
    fetchJSON<Paginated<Album>>(`${BASE}/albums?page=${page}&limit=${limit}`),
  getAlbum: (id: number) => fetchJSON<Album>(`${BASE}/albums/${id}`),
  updateAlbum: (id: number, data: { note?: string; favorite?: boolean }) =>
    patchJSON<{ ok: boolean }>(`${BASE}/albums/${id}`, data),
  getAlbumTracks: (id: number) =>
    fetchJSON<{ items: Track[] }>(`${BASE}/albums/${id}/tracks`),

  listTracks: (page = 1, limit = 50) =>
    fetchJSON<Paginated<Track>>(`${BASE}/tracks?page=${page}&limit=${limit}`),
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
};
