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
            </Routes>
          </PageWrapper>
        </main>
      </BrowserRouter>
    </ThemeProvider>
  );
}
