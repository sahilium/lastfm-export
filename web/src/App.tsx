import { BrowserRouter, Routes, Route, Link, useLocation } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { ThemeProvider, useTheme } from './lib/theme';
import ArtistsPage from './pages/ArtistsPage';
import ArtistPage from './pages/ArtistPage';
import AlbumsPage from './pages/AlbumsPage';
import AlbumPage from './pages/AlbumPage';
import TracksPage from './pages/TracksPage';
import TrackPage from './pages/TrackPage';
import SearchPage from './pages/SearchPage';
import CollectionsPage from './pages/CollectionsPage';
import CollectionPage from './pages/CollectionPage';
import SettingsPage from './pages/SettingsPage';
import './App.css';

function ScrollToTop() {
  const { pathname } = useLocation();
  useEffect(() => { window.scrollTo(0, 0); }, [pathname]);
  return null;
}

function PageWrapper({ children }: { children: React.ReactNode }) {
  const { pathname } = useLocation();
  return (
    <div key={pathname} className="page-enter">
      {children}
    </div>
  );
}

function Nav() {
  const { pathname } = useLocation();
  const { theme, toggle } = useTheme();
  const [open, setOpen] = useState(false);

  useEffect(() => { setOpen(false); }, [pathname]);

  const links = [
    { to: '/', label: 'Artists' },
    { to: '/albums', label: 'Albums' },
    { to: '/tracks', label: 'Tracks' },
    { to: '/collections', label: 'Collections' },
    { to: '/search', label: 'Search' },
  ];

  const isActive = (to: string) =>
    to === '/' ? pathname === '/' : pathname.startsWith(to);

  return (
    <nav className="nav">
      <div className="nav-inner">
        <Link to="/" className="nav-brand">musiclib</Link>

        <div className="nav-desktop">
          {links.map(l => (
            <Link key={l.to} to={l.to} className={`nav-link ${isActive(l.to) ? 'active' : ''}`}>
              {l.label}
            </Link>
          ))}
        </div>

        <div className="nav-right">
          <button className="theme-toggle" onClick={toggle} title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} theme`}>
            {theme === 'dark' ? '\u2600\uFE0F' : '\uD83C\uDF19'}
          </button>
          <Link
            to="/settings"
            className={`nav-cog ${pathname === '/settings' ? 'active' : ''}`}
            title="Settings"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="12" cy="12" r="3" />
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
            </svg>
          </Link>
          <button className="nav-hamburger" onClick={() => setOpen(!open)} aria-label="Menu">
            <span className={`hamburger-line ${open ? 'open' : ''}`}></span>
          </button>
        </div>
      </div>

      {open && (
        <div className="nav-mobile">
          {links.map(l => (
            <Link key={l.to} to={l.to} className={`nav-mobile-link ${isActive(l.to) ? 'active' : ''}`}>
              {l.label}
            </Link>
          ))}
          <Link to="/settings" className={`nav-mobile-link ${pathname === '/settings' ? 'active' : ''}`}>
            Settings
          </Link>
        </div>
      )}
    </nav>
  );
}

export default function App() {
  return (
    <ThemeProvider>
      <BrowserRouter>
        <ScrollToTop />
        <Nav />
        <main className="container">
          <PageWrapper>
            <Routes>
              <Route path="/" element={<ArtistsPage />} />
              <Route path="/artists/:id" element={<ArtistPage />} />
              <Route path="/albums" element={<AlbumsPage />} />
              <Route path="/albums/:id" element={<AlbumPage />} />
              <Route path="/tracks" element={<TracksPage />} />
              <Route path="/tracks/:id" element={<TrackPage />} />
              <Route path="/search" element={<SearchPage />} />
              <Route path="/collections" element={<CollectionsPage />} />
              <Route path="/collections/:id" element={<CollectionPage />} />
              <Route path="/settings" element={<SettingsPage />} />
            </Routes>
          </PageWrapper>
        </main>
      </BrowserRouter>
    </ThemeProvider>
  );
}
