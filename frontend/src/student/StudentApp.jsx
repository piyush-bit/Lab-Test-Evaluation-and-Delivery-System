import React from 'react';
import { HashRouter, Routes, Route, useLocation, useNavigate } from 'react-router-dom';
import { StudentProvider, useStudent } from './StudentContext';
import Welcome from './pages/Welcome';
import ActiveWorkspace from './pages/ActiveWorkspace';
import CommandPalette from './components/CommandPalette';

function StudentAppContent() {
  const { 
    activeWorkspacePath, 
    setActiveWorkspacePath,
    runMode,
    setRunMode,
    runStatus,
    handleRunTests
  } = useStudent();
  const location = useLocation();
  const navigate = useNavigate();
  const [dropdownOpen, setDropdownOpen] = React.useState(false);

  const hasActiveWorkspace = !!activeWorkspacePath;
  const isWorkspaceRoute = location.pathname === '/workspace';

  // Get exercise folder name
  const exerciseName = activeWorkspacePath 
    ? activeWorkspacePath.split('/').pop().split('\\').pop()
    : '';

  // Close dropdown on click outside
  React.useEffect(() => {
    if (!dropdownOpen) return;
    const handleOutsideClick = () => setDropdownOpen(false);
    document.addEventListener('click', handleOutsideClick);
    return () => document.removeEventListener('click', handleOutsideClick);
  }, [dropdownOpen]);

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
            <div style={{ display: 'flex', alignItems: 'center', position: 'relative', height: '24px' }}>
              {/* Left Part: Combined Run Trigger Button */}
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  handleRunTests();
                }}
                disabled={runStatus === 'running'}
                title={runMode === 'docker' ? 'Execute public tests in Docker container' : 'Execute public tests locally on host'}
                style={{
                  backgroundColor: 'var(--btn-primary-bg)',
                  border: 'none',
                  borderTopLeftRadius: '6px',
                  borderBottomLeftRadius: '6px',
                  padding: '4px 12px',
                  cursor: runStatus === 'running' ? 'not-allowed' : 'pointer',
                  color: 'var(--btn-primary-text)',
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: '6px',
                  fontSize: '0.72rem',
                  fontWeight: 600,
                  opacity: runStatus === 'running' ? 0.75 : 1,
                  transition: 'opacity 0.15s ease',
                  height: '100%',
                  borderRight: '1px solid rgba(255, 255, 255, 0.15)'
                }}
              >
                {runStatus === 'running' ? (
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" style={{ animation: 'spin 1s linear infinite' }}>
                    <circle cx="12" cy="12" r="10" strokeDasharray="30" strokeDashoffset="10" />
                  </svg>
                ) : runMode === 'docker' ? (
                  // Play with Docker containers stacked boxes icon
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M3 2v14l10-7-10-7z" />
                    <rect x="15" y="14" width="4" height="3" fill="currentColor" rx="0.5" />
                    <rect x="20" y="14" width="4" height="3" fill="currentColor" rx="0.5" />
                    <rect x="17.5" y="10" width="4" height="3" fill="currentColor" rx="0.5" />
                  </svg>
                ) : (
                  // Play with Local Monitor icon
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M3 2v14l10-7-10-7z" />
                    <path d="M15 11h8v6h-8v-6zm3 7h2v1.5h-2V18z" />
                  </svg>
                )}
                <span>{runStatus === 'running' ? 'Running' : `Run (${runMode === 'docker' ? 'Docker' : 'Local'})`}</span>
              </button>

              {/* Right Part: Dropdown arrow */}
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  setDropdownOpen(!dropdownOpen);
                }}
                disabled={runStatus === 'running'}
                style={{
                  backgroundColor: 'var(--btn-primary-bg)',
                  border: 'none',
                  borderTopRightRadius: '6px',
                  borderBottomRightRadius: '6px',
                  padding: '4px 8px',
                  cursor: runStatus === 'running' ? 'not-allowed' : 'pointer',
                  color: 'var(--btn-primary-text)',
                  display: 'inline-flex',
                  alignItems: 'center',
                  height: '100%',
                  opacity: runStatus === 'running' ? 0.75 : 1
                }}
              >
                <svg width="8" height="8" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3.5" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="6 9 12 15 18 9" />
                </svg>
              </button>

              {/* Dropdown Menu Overlay */}
              {dropdownOpen && (
                <div style={{
                  position: 'absolute',
                  top: '100%',
                  right: 0,
                  marginTop: '6px',
                  backgroundColor: 'var(--bg-card)',
                  border: '1px solid var(--border-color)',
                  borderRadius: '8px',
                  boxShadow: '0 8px 24px rgba(0,0,0,0.18)',
                  padding: '6px',
                  zIndex: 20,
                  minWidth: '220px',
                  display: 'flex',
                  flexDirection: 'column',
                  gap: '4px',
                  textAlign: 'left'
                }}>
                  <div style={{
                    fontSize: '0.62rem',
                    fontWeight: 700,
                    color: 'var(--text-muted)',
                    textTransform: 'uppercase',
                    letterSpacing: '0.05em',
                    padding: '4px 8px',
                    borderBottom: '1px solid var(--border-color)',
                    marginBottom: '4px'
                  }}>
                    Select & Run Environment
                  </div>

                  {/* Docker Environment Option */}
                  <div
                    onClick={(e) => {
                      e.stopPropagation();
                      setRunMode('docker');
                      setDropdownOpen(false);
                      handleRunTests('', 'docker');
                    }}
                    style={{
                      padding: '8px 10px',
                      borderRadius: '6px',
                      cursor: 'pointer',
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                      backgroundColor: runMode === 'docker' ? 'rgba(79, 70, 229, 0.06)' : 'transparent',
                      transition: 'background-color 0.15s ease'
                    }}
                  >
                    <div>
                      <div style={{ fontSize: '0.75rem', fontWeight: 600, color: runMode === 'docker' ? 'var(--primary)' : 'var(--text-primary)', display: 'flex', alignItems: 'center', gap: '4px' }}>
                        🐳 Docker Container
                      </div>
                      <div style={{ fontSize: '0.65rem', color: 'var(--text-muted)', marginTop: '2px' }}>
                        Run in secure isolated sandbox
                      </div>
                    </div>
                    {runMode === 'docker' && (
                      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="var(--primary)" strokeWidth="3.5" strokeLinecap="round" strokeLinejoin="round">
                        <polyline points="20 6 9 17 4 12" />
                      </svg>
                    )}
                  </div>

                  {/* Local Host Environment Option */}
                  <div
                    onClick={(e) => {
                      e.stopPropagation();
                      setRunMode('local');
                      setDropdownOpen(false);
                      handleRunTests('', 'local');
                    }}
                    style={{
                      padding: '8px 10px',
                      borderRadius: '6px',
                      cursor: 'pointer',
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                      backgroundColor: runMode === 'local' ? 'rgba(79, 70, 229, 0.06)' : 'transparent',
                      transition: 'background-color 0.15s ease'
                    }}
                  >
                    <div>
                      <div style={{ fontSize: '0.75rem', fontWeight: 600, color: runMode === 'local' ? 'var(--primary)' : 'var(--text-primary)', display: 'flex', alignItems: 'center', gap: '4px' }}>
                        💻 Local Host
                      </div>
                      <div style={{ fontSize: '0.65rem', color: 'var(--text-muted)', marginTop: '2px' }}>
                        Run on native host machine
                      </div>
                    </div>
                    {runMode === 'local' && (
                      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="var(--primary)" strokeWidth="3.5" strokeLinecap="round" strokeLinejoin="round">
                        <polyline points="20 6 9 17 4 12" />
                      </svg>
                    )}
                  </div>
                </div>
              )}
            </div>
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
