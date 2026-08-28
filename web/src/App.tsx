import { BrowserRouter, Routes, Route, Link, useLocation } from 'react-router-dom';
import ArtistsPage from './pages/ArtistsPage';
import ArtistPage from './pages/ArtistPage';
import AlbumsPage from './pages/AlbumsPage';
import AlbumPage from './pages/AlbumPage';
import TracksPage from './pages/TracksPage';
import TrackPage from './pages/TrackPage';
import SearchPage from './pages/SearchPage';
import './App.css';

function Nav() {
  const { pathname } = useLocation();
  const links = [
    { to: '/', label: 'Artists' },
    { to: '/albums', label: 'Albums' },
    { to: '/tracks', label: 'Tracks' },
    { to: '/search', label: 'Search' },
  ];

  return (
    <nav className="nav">
      <div className="nav-inner">
        <Link to="/" className="nav-brand">musiclib</Link>
        <div className="nav-links">
          {links.map(l => (
            <Link
              key={l.to}
              to={l.to}
              className={`nav-link ${pathname === l.to || (l.to !== '/' && pathname.startsWith(l.to)) ? 'active' : ''}`}
            >
              {l.label}
            </Link>
          ))}
        </div>
      </div>
    </nav>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <Nav />
      <main className="container">
        <Routes>
          <Route path="/" element={<ArtistsPage />} />
          <Route path="/artists/:id" element={<ArtistPage />} />
          <Route path="/albums" element={<AlbumsPage />} />
          <Route path="/albums/:id" element={<AlbumPage />} />
          <Route path="/tracks" element={<TracksPage />} />
          <Route path="/tracks/:id" element={<TrackPage />} />
          <Route path="/search" element={<SearchPage />} />
        </Routes>
      </main>
    </BrowserRouter>
  );
}
