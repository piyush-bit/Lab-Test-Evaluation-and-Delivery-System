import React, { useState } from 'react';
import { useStudent } from '../StudentContext';
import { TrashIcon, RefreshIcon, PlusIcon } from './Icons';

export default function SystemStatus() {
  const {
    daemonConnected,
    dockerRunning,
    isDarkMode,
    setIsDarkMode,
    remoteServers,
    remoteServerStatuses,
    addRemoteServer,
    removeRemoteServer,
    checkAllRemoteServersHealth
  } = useStudent();

  const [newServerUrl, setNewServerUrl] = useState('');
  const [isAdding, setIsAdding] = useState(false);

  const handleAddSubmit = (e) => {
    e.preventDefault();
    if (!newServerUrl.trim()) return;
    addRemoteServer(newServerUrl.trim());
    setNewServerUrl('');
    setIsAdding(false);
  };

  return (
    <div className="vscode-right-col">
      {/* System Status Section */}
      <div className="vscode-section">
        <h2>System Status</h2>
        <div className="vscode-technical-status-stack">
          <div className="tech-status-row">
            <span className="tech-label">DAEMON:</span>
            <span className={`tech-value status-${daemonConnected ? 'healthy' : 'dead'}`}>
              {daemonConnected ? 'ONLINE' : 'OFFLINE'}
            </span>
          </div>

          <div className="tech-status-row">
            <span className="tech-label">DOCKER:</span>
            <span className={`tech-value status-${dockerRunning ? 'healthy' : 'warn'}`}>
              {dockerRunning ? 'RUNNING' : 'STOPPED'}
            </span>
          </div>
        </div>
      </div>

      {/* Remote Servers Management Section */}
      <div className="vscode-section" style={{ marginTop: '28px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
          <h2 style={{ margin: 0 }}>Remote Servers</h2>
          <div style={{ display: 'flex', gap: '6px' }}>
            <button
              onClick={() => checkAllRemoteServersHealth()}
              title="Refresh all server statuses"
              className="vscode-theme-text-btn"
              style={{ padding: '3px 6px', display: 'inline-flex', alignItems: 'center', gap: '4px', cursor: 'pointer' }}
            >
              <RefreshIcon size={12} />
            </button>
            <button
              onClick={() => setIsAdding(!isAdding)}
              title="Add Remote Server"
              className="vscode-theme-text-btn"
              style={{ padding: '3px 6px', display: 'inline-flex', alignItems: 'center', gap: '4px', cursor: 'pointer' }}
            >
              <PlusIcon size={12} />
            </button>
          </div>
        </div>

        {isAdding && (
          <form onSubmit={handleAddSubmit} style={{ marginBottom: '12px', display: 'flex', gap: '6px' }}>
            <input
              type="text"
              placeholder="http://localhost:8080"
              value={newServerUrl}
              onChange={(e) => setNewServerUrl(e.target.value)}
              style={{
                flex: 1,
                padding: '4px 8px',
                fontSize: '0.75rem',
                fontFamily: 'var(--font-mono)',
                backgroundColor: 'var(--bg-main)',
                border: '1px solid var(--border-color)',
                borderRadius: '4px',
                color: 'var(--text-primary)',
                outline: 'none'
              }}
              autoFocus
            />
            <button
              type="submit"
              className="vscode-theme-text-btn"
              style={{ backgroundColor: 'var(--primary)', color: '#fff', border: 'none', cursor: 'pointer' }}
            >
              Add
            </button>
          </form>
        )}

        <div className="vscode-technical-status-stack">
          {remoteServers.length === 0 ? (
            <div style={{ fontSize: '0.78rem', color: 'var(--text-muted)', fontStyle: 'italic' }}>
              No remote servers configured.
            </div>
          ) : (
            remoteServers.map((serverUrl) => {
              const statusInfo = remoteServerStatuses[serverUrl] || { loading: true };
              const isOnline = statusInfo.online;
              const isLoading = statusInfo.loading;

              return (
                <div key={serverUrl} className="tech-status-row" style={{ justifyContent: 'space-between' }}>
                  <div style={{ display: 'flex', flexDirection: 'column', overflow: 'hidden', marginRight: '6px' }}>
                    <span style={{ fontSize: '0.78rem', fontFamily: 'var(--font-mono)', color: 'var(--text-primary)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }} title={serverUrl}>
                      {serverUrl}
                    </span>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px', flexShrink: 0 }}>
                    <span className={`tech-value status-${isLoading ? 'warn' : isOnline ? 'healthy' : 'dead'}`} style={{ fontSize: '0.72rem' }}>
                      {isLoading ? 'CHECKING...' : isOnline ? 'ONLINE' : 'OFFLINE'}
                    </span>
                    <button
                      onClick={() => removeRemoteServer(serverUrl)}
                      title="Remove server"
                      style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-muted)', padding: '2px', display: 'flex', alignItems: 'center' }}
                    >
                      <TrashIcon size={12} />
                    </button>
                  </div>
                </div>
              );
            })
          )}
        </div>
      </div>

      {/* Appearance Section */}
      <div className="vscode-section" style={{ marginTop: '28px' }}>
        <h2>Appearance</h2>
        <div className="vscode-technical-status-stack">
          <div className="tech-status-row">
            <span className="tech-label">THEME:</span>
            <button 
              className="vscode-theme-text-btn" 
              onClick={() => setIsDarkMode(!isDarkMode)}
            >
              {isDarkMode ? 'DARK' : 'LIGHT'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
