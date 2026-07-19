import React from 'react';
import { HashRouter, Routes, Route, useLocation, useNavigate } from 'react-router-dom';
import { StudentProvider, useStudent } from './StudentContext';
import Welcome from './pages/Welcome';
import ActiveWorkspace from './pages/ActiveWorkspace';
import CommandPalette from './components/CommandPalette';

function StudentAppContent() {
  const { activeWorkspacePath, setActiveWorkspacePath } = useStudent();
  const location = useLocation();
  const navigate = useNavigate();

  const hasActiveWorkspace = !!activeWorkspacePath;
  const isWorkspaceRoute = location.pathname === '/workspace';

  // Get exercise folder name
  const exerciseName = activeWorkspacePath 
    ? activeWorkspacePath.split('/').pop().split('\\').pop()
    : '';

  return (
    <div className={`app-viewport ${hasActiveWorkspace ? 'workspace-active' : ''}`}>
      {/* Background Grid Lines (only visible when workspace is NOT active) */}
      {!hasActiveWorkspace && (
        <div className="grid-lines-container">
          <div className="vertical-line v-line-1"></div>
          <div className="vertical-line v-line-2"></div>
          <div className="vertical-line v-line-3"></div>
        </div>
      )}

      {/* Global Header Bar - visible when workspace is active */}
      {hasActiveWorkspace && (
        <div className="workspace-top-bar" style={{
          height: '40px',
          width: '100%',
          borderBottom: '1px solid var(--border-color)',
          backgroundColor: 'var(--bg-card)',
          backdropFilter: 'var(--glass-blur)',
          WebkitBackdropFilter: 'var(--glass-blur)',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          padding: '0 16px',
          zIndex: 10,
          userSelect: 'none'
        }}>
          {/* Left: Section Label */}
          <div style={{
            fontSize: '0.75rem',
            fontWeight: 700,
            color: 'var(--text-primary)',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            padding: '4px 8px'
          }}>
            Editor
          </div>

          {/* Middle: Exercise Name (name of the directory we opened) */}
          <div style={{
            fontSize: '0.8rem',
            fontWeight: 600,
            color: 'var(--text-primary)',
            fontFamily: 'var(--font-sans)',
            display: 'flex',
            alignItems: 'center',
            gap: '8px'
          }}>
            <span>{exerciseName}</span>
          </div>

          {/* Right: Two future scope buttons + Close Workspace */}
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <button
              title="Run Tests (Future Scope)"
              style={{
                background: 'none',
                border: '1px solid var(--border-color)',
                backgroundColor: 'rgba(0,0,0,0.02)',
                borderRadius: '6px',
                padding: '4px 8px',
                cursor: 'pointer',
                color: 'var(--text-secondary)',
                display: 'flex',
                alignItems: 'center',
                gap: '4px',
                fontSize: '0.72rem',
                fontWeight: 600
              }}
            >
              <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{ color: 'var(--accent)' }}>
                <polygon points="5 3 19 12 5 21 5 3"/>
              </svg>
              Run
            </button>
            <button
              title="Submit Answer (Future Scope)"
              style={{
                background: 'none',
                border: '1px solid var(--border-color)',
                backgroundColor: 'rgba(0,0,0,0.02)',
                borderRadius: '6px',
                padding: '4px 8px',
                cursor: 'pointer',
                color: 'var(--text-secondary)',
                display: 'flex',
                alignItems: 'center',
                gap: '4px',
                fontSize: '0.72rem',
                fontWeight: 600
              }}
            >
              <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{ color: 'var(--primary)' }}>
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Submit
            </button>
            <span style={{ fontSize: '0.8rem', color: 'var(--border-color)', margin: '0 2px' }}>|</span>
            <button
              title="Close Workspace"
              onClick={() => {
                setActiveWorkspacePath('');
                navigate('/');
              }}
              style={{
                background: 'none',
                border: 'none',
                cursor: 'pointer',
                color: 'var(--accent-red)',
                display: 'flex',
                alignItems: 'center',
                padding: '4px',
                borderRadius: '4px'
              }}
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
                <line x1="9" y1="9" x2="15" y2="15" />
                <line x1="15" y1="9" x2="9" y2="15" />
              </svg>
            </button>
          </div>
        </div>
      )}

      {/* Main Content Area */}
      <main className={`main-viewport-content ${hasActiveWorkspace ? 'workspace-active' : ''} ${isWorkspaceRoute ? 'workspace-route-active' : ''}`}>
        <Routes>
          <Route path="/" element={<Welcome />} />
          <Route path="/workspace" element={<ActiveWorkspace />} />
        </Routes>
      </main>

      {/* Command Palette Overlay */}
      <CommandPalette />
    </div>
  );
}

export default function StudentApp() {
  return (
    <HashRouter>
      <StudentProvider>
        <StudentAppContent />
      </StudentProvider>
    </HashRouter>
  );
}
