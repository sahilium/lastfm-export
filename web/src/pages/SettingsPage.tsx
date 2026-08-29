import { useState, useEffect, useCallback } from 'react';

interface Settings {
  mode: 'local' | 'turso' | 'unconfigured';
  turso_url: string;
  has_token: boolean;
  turso_token: string;
  local_mode: boolean;
}

export default function SettingsPage() {
  const [settings, setSettings] = useState<Settings | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch('/api/settings');
      const data = await res.json();
      setSettings(data);
    } catch (e) {
      console.error('Failed to load settings', e);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  if (loading) return <div className="loading">Loading...</div>;
  if (!settings) return <div className="error">Failed to load settings</div>;

  return (
    <div>
      <div className="page-header">
        <h1>Settings</h1>
      </div>

      <section className="section">
        <h2>Database Connection</h2>

        <div className="settings-status">
          <div className="settings-status-row">
            <span className="settings-status-label">Mode</span>
            <span className={`settings-mode-badge ${settings.mode}`}>
              {settings.mode === 'local' && 'Local SQLite'}
              {settings.mode === 'turso' && 'Turso (cloud)'}
              {settings.mode === 'unconfigured' && 'Not configured'}
            </span>
          </div>

          {settings.mode === 'local' && (
            <div className="settings-status-row">
              <span className="settings-status-label">Source</span>
              <span className="settings-status-value">
                Started with <code>--db</code> flag
              </span>
            </div>
          )}

          {settings.mode === 'turso' && (
            <>
              <div className="settings-status-row">
                <span className="settings-status-label">Turso URL</span>
                <span className="settings-status-value settings-url">
                  {settings.turso_url}
                </span>
              </div>
              <div className="settings-status-row">
                <span className="settings-status-label">Auth Token</span>
                <span className="settings-status-value">
                  {settings.has_token ? (
                    <span className="settings-token-ok">{settings.turso_token}</span>
                  ) : (
                    <em className="text-secondary">not set</em>
                  )}
                </span>
              </div>
            </>
          )}

          {settings.mode === 'unconfigured' && (
            <div className="settings-notice">
              No database configured. Either start with <code>musiclib serve --db &lt;path&gt;</code> for local mode, or configure Turso below and restart.
            </div>
          )}
        </div>
      </section>

      <section className="section">
        <h2>How to change database</h2>
        <div className="settings-help">
          <div className="settings-help-item">
            <h3>Local mode</h3>
            <p>Start the server with a local SQLite database:</p>
            <pre className="settings-code">musiclib serve --db /path/to/musiclib.db</pre>
            <p>Data features (artists, albums, tracks, collections) read from and write to this file.</p>
          </div>

          <div className="settings-help-item">
            <h3>Turso (cloud) mode</h3>
            <p>Start without <code>--db</code>, then configure Turso in the settings page and restart:</p>
            <pre className="settings-code">musiclib serve</pre>
            <p>Go to this page, enter your Turso database URL and auth token, then restart the server.</p>
          </div>
        </div>
      </section>
    </div>
  );
}
