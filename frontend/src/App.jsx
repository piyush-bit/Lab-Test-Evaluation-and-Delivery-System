import React, { useState, useEffect, useRef } from 'react';
import './App.css';

// SVG Icons matching VS Code look
const OpenFolderIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>
);

const CreateIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="12" y1="18" x2="12" y2="12"/><line x1="9" y1="15" x2="15" y2="15"/></svg>
);

const HistoryIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
);

const BackIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="19" y1="12" x2="5" y2="12"/><polyline points="12 19 5 12 12 5"/></svg>
);

const CheckIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"/></svg>
);

const CloseIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
);

export default function App() {
  const [isDarkMode, setIsDarkMode] = useState(true);
  const [screen, setScreen] = useState('welcome'); // welcome, create-workspace, active-workspace

  // Server & Daemon Status
  const [daemonConnected, setDaemonConnected] = useState(true);
  const [dockerRunning, setDockerRunning] = useState(false);
  const [currentCwd, setCurrentCwd] = useState('');

  // VS Code-style Command Palette folder picker
  const [showQuickOpen, setShowQuickOpen] = useState(false);
  const [quickOpenPath, setQuickOpenPath] = useState('');
  const [quickOpenDirs, setQuickOpenDirs] = useState([]);
  const [quickOpenParent, setQuickOpenParent] = useState('');
  const [quickOpenCallback, setQuickOpenCallback] = useState(null);

  // Recents Storage
  const [recents, setRecents] = useState(() => {
    try {
      const stored = localStorage.getItem('recent_workspaces');
      return stored ? JSON.parse(stored) : [
        '~/Developer/TDES/workspace/go101-lab01',
        '~/Developer/TDES/workspace/python-basics'
      ];
    } catch {
      return [];
    }
  });

  // Create Workspace inputs
  const [labID, setLabID] = useState('go101-lab01');
  const [version, setVersion] = useState('v1.0');
  const [remoteURL, setRemoteURL] = useState('http://localhost:8080');
  const [orgID, setOrgID] = useState('default');
  const [targetPath, setTargetPath] = useState('');

  // Active Workspace State
  const [activeWorkspacePath, setActiveWorkspacePath] = useState('');
  const [validationError, setValidationError] = useState('');
  const [loading, setLoading] = useState(false);

  const quickOpenRef = useRef(null);
  const lastFetchedBaseDirRef = useRef('');

  // Sync Dark/Light Themes
  useEffect(() => {
    const root = document.documentElement;
    if (isDarkMode) {
      root.classList.add('dark');
    } else {
      root.classList.remove('dark');
    }
  }, [isDarkMode]);

  // Daemon Connection Status Heartbeat Check
  useEffect(() => {
    const fetchStatus = async () => {
      try {
        const res = await fetch('/api/status');
        if (res.ok) {
          const data = await res.json();
          setDaemonConnected(true);
          setDockerRunning(data.docker_running);
          if (data.workspace && !currentCwd) {
            setCurrentCwd(data.workspace);
            setTargetPath(data.workspace);
          }
        } else {
          setDaemonConnected(false);
        }
      } catch {
        setDaemonConnected(false);
      }
    };

    fetchStatus();
    const interval = setInterval(fetchStatus, 3000);
    return () => clearInterval(interval);
  }, [currentCwd]);

  // Command Palette Click Outside & ESC listeners
  useEffect(() => {
    function handleClickOutside(event) {
      if (quickOpenRef.current && !quickOpenRef.current.contains(event.target)) {
        setShowQuickOpen(false);
      }
    }
    function handleEscKey(event) {
      if (event.key === 'Escape') {
        setShowQuickOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    document.addEventListener("keydown", handleEscKey);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("keydown", handleEscKey);
    };
  }, []);

  // Path parts parser for filtering typed queries
  const getPathParts = (path) => {
    const sep = path.includes('/') ? '/' : '\\';
    const lastIdx = path.lastIndexOf(sep);
    if (lastIdx === -1) {
      return { baseDir: path, filterQuery: '', sep };
    }
    const baseDir = path.substring(0, lastIdx + 1); // includes the last separator
    const filterQuery = path.substring(lastIdx + 1);
    return { baseDir, filterQuery, sep };
  };

  // Fetch folders for Command Palette
  const fetchQuickOpenDirs = async (path) => {
    try {
      const res = await fetch(`/api/browse?path=${encodeURIComponent(path)}`);
      if (res.ok) {
        const data = await res.json();
        setQuickOpenDirs(data.directories || []);
        setQuickOpenParent(data.parent_path || '');
      }
    } catch {
      // silent fail
    }
  };

  // Detect baseDir transitions to fetch directories, ignoring typing filtering queries
  useEffect(() => {
    if (!showQuickOpen) return;

    const { baseDir } = getPathParts(quickOpenPath);
    if (baseDir && baseDir !== lastFetchedBaseDirRef.current) {
      lastFetchedBaseDirRef.current = baseDir;
      fetchQuickOpenDirs(baseDir);
    }
  }, [quickOpenPath, showQuickOpen]);

  // Trigger folder picker modal (Quick Open style)
  const triggerQuickOpen = (initialPath, callbackFn) => {
    const startPath = initialPath || currentCwd || '~/';
    setQuickOpenPath(startPath);
    
    const { baseDir } = getPathParts(startPath);
    lastFetchedBaseDirRef.current = baseDir;
    fetchQuickOpenDirs(baseDir);
    
    setQuickOpenCallback(() => callbackFn);
    setShowQuickOpen(true);
  };

  const handleQuickOpenNavigate = (newPath) => {
    setQuickOpenPath(newPath);
    const { baseDir } = getPathParts(newPath);
    lastFetchedBaseDirRef.current = baseDir;
    fetchQuickOpenDirs(baseDir);
  };

  const handleGoUp = () => {
    if (quickOpenParent) {
      const parentPath = quickOpenParent;
      const sep = parentPath.includes('/') ? '/' : '\\';
      const cleanParent = parentPath.endsWith(sep) ? parentPath : parentPath + sep;
      handleQuickOpenNavigate(cleanParent);
    }
  };

  const handleConfirmQuickOpen = () => {
    if (quickOpenCallback) {
      quickOpenCallback(quickOpenPath);
    }
    setShowQuickOpen(false);
  };

  // Keyboard navigation inside Command Palette input
  const handleQuickOpenKeyDown = (e) => {
    if (e.key === 'Enter') {
      handleConfirmQuickOpen();
    }
  };

  const addWorkspaceToRecents = (path) => {
    const updated = [path, ...recents.filter(r => r !== path)].slice(0, 5);
    setRecents(updated);
    localStorage.setItem('recent_workspaces', JSON.stringify(updated));
  };

  // Open Workspace Action
  const handleOpenWorkspace = async (path) => {
    setValidationError('');
    setLoading(true);

    try {
      const res = await fetch(`/api/validate-workspace?path=${encodeURIComponent(path)}`);
      if (res.ok) {
        const data = await res.json();
        if (data.valid) {
          setActiveWorkspacePath(data.path);
          addWorkspaceToRecents(data.path);
          setScreen('active-workspace');
        } else {
          setValidationError(data.error || "No manifest.json found. Ensure this is an initialized TDES directory.");
        }
      } else {
        setValidationError("Failed to communicate with local agent.");
      }
    } catch (err) {
      setValidationError("Connection error: " + err.message);
    }
    setLoading(false);
  };

  // Create Workspace Action
  const handleCreateWorkspace = async () => {
    setValidationError('');
    setLoading(true);

    try {
      const fetchRes = await fetch('/api/fetch', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          exercise_id: labID,
          version: version,
          remote_url: remoteURL,
          org_id: orgID
        })
      });
      
      if (!fetchRes.ok) {
        const fetchErr = await fetchRes.json();
        setValidationError(`Fetch failed: ${fetchErr.error || 'Server error'}`);
        setLoading(false);
        return;
      }

      const initRes = await fetch('/api/init', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          exercise_id: labID,
          version: version,
          target_dir: targetPath
        })
      });

      if (!initRes.ok) {
        const initErr = await initRes.json();
        setValidationError(`Init failed: ${initErr.error || 'Server error'}`);
        setLoading(false);
        return;
      }

      setActiveWorkspacePath(targetPath);
      addWorkspaceToRecents(targetPath);
      setScreen('active-workspace');
    } catch (err) {
      setValidationError("Connection error: " + err.message);
    }
    setLoading(false);
  };

  // Filter visible subdirectories on typed text
  const { filterQuery } = getPathParts(quickOpenPath);
  const filteredDirs = quickOpenDirs.filter(dir => 
    dir.toLowerCase().includes(filterQuery.toLowerCase())
  );

  return (
    <div className="app-viewport">
      {/* Background Grid Lines (Teachyst Style) */}
      <div className="grid-lines-container">
        <div className="vertical-line v-line-1"></div>
        <div className="vertical-line v-line-2"></div>
        <div className="vertical-line v-line-3"></div>
      </div>

      {/* Main Content Area */}
      <main className="main-viewport-content">

        {/* SCREEN 1: Welcome Screen (VS Code Columns style) */}
        {screen === 'welcome' && (
          <div className="vscode-welcome-container">
            <header className="vscode-header">
              <h1>TDES Console</h1>
              <p>Lab evaluation and packaging dashboard</p>
              {validationError && (
                <div className="validation-alert-error" style={{ marginTop: '14px' }}>
                  {validationError}
                </div>
              )}
            </header>

            <div className="vscode-columns-grid">
              
              {/* LEFT COLUMN: Start & Recents */}
              <div className="vscode-left-col">
                <div className="vscode-section">
                  <h2>Start</h2>
                  <div className="vscode-action-list">
                    <div className="vscode-action-item" onClick={() => {
                      setValidationError('');
                      triggerQuickOpen(targetPath || currentCwd, (path) => {
                        handleOpenWorkspace(path);
                      });
                    }}>
                      <span className="action-icon"><OpenFolderIcon /></span>
                      <div className="action-details">
                        <span className="action-title">Open Folder...</span>
                        <span className="action-desc">Open an already initialized exercise directory</span>
                      </div>
                    </div>

                    <div className="vscode-action-item" onClick={() => {
                      setValidationError('');
                      setScreen('create-workspace');
                    }}>
                      <span className="action-icon"><CreateIcon /></span>
                      <div className="action-details">
                        <span className="action-title">Initialize Lab...</span>
                        <span className="action-desc">Fetch and initialize a fresh template package</span>
                      </div>
                    </div>
                  </div>
                </div>

                <div className="vscode-section" style={{ marginTop: '28px' }}>
                  <h2>Recent</h2>
                  <div className="vscode-recents-list">
                    {recents.length === 0 ? (
                      <div className="empty-recents-msg">No recent folders opened.</div>
                    ) : (
                      recents.map((path, idx) => (
                        <div 
                          key={idx} 
                          className="vscode-recent-item"
                          onClick={() => handleOpenWorkspace(path)}
                        >
                          <span className="recent-icon"><HistoryIcon /></span>
                          <div className="recent-details">
                            <span className="recent-name">{path.split('/').pop() || path.split('\\').pop()}</span>
                            <span className="recent-path">{path}</span>
                          </div>
                        </div>
                      ))
                    )}
                  </div>
                </div>
              </div>

              {/* RIGHT COLUMN: Status & Appearance */}
              <div className="vscode-right-col">
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

            </div>
          </div>
        )}

        {/* SCREEN 2: Create Workspace View */}
        {screen === 'create-workspace' && (
          <div className="action-form-container glass-card">
            <button className="back-btn" onClick={() => setScreen('welcome')}>
              <BackIcon /> Back
            </button>
            <h2 className="form-title">Initialize Exercise</h2>
            <p className="form-description">Enter details to retrieve the lab boilerplate and configure files.</p>

            <div className="form-grid">
              <div className="form-group-wrap">
                <label className="field-label">Lab/Exercise ID</label>
                <input 
                  type="text" 
                  value={labID} 
                  onChange={e => setLabID(e.target.value)} 
                  placeholder="e.g. go101-lab01"
                />
              </div>

              <div className="form-group-wrap">
                <label className="field-label">Version</label>
                <input 
                  type="text" 
                  value={version} 
                  onChange={e => setVersion(e.target.value)} 
                  placeholder="e.g. v1.0"
                />
              </div>
            </div>

            <div className="form-grid">
              <div className="form-grid">
                <div className="form-group-wrap">
                  <label className="field-label">Organization ID</label>
                  <input 
                    type="text" 
                    value={orgID} 
                    onChange={e => setOrgID(e.target.value)} 
                  />
                </div>
              </div>

              <div className="form-group-wrap">
                <label className="field-label">Registry Server URL</label>
                <input 
                  type="text" 
                  value={remoteURL} 
                  onChange={e => setRemoteURL(e.target.value)} 
                />
              </div>
            </div>

            <div className="form-group-wrap">
              <label className="field-label">Target Folder Path</label>
              <div className="vscode-input-wrapper">
                <input 
                  type="text" 
                  value={targetPath}
                  onChange={e => setTargetPath(e.target.value)}
                  placeholder="Absolute directory where lab will be scaffolded"
                  className="vscode-input"
                />
                <button 
                  className="browse-folder-btn" 
                  onClick={() => triggerQuickOpen(targetPath, (path) => {
                    setTargetPath(path);
                  })}
                  title="Browse folders"
                >
                  <OpenFolderIcon />
                </button>
              </div>
            </div>

            {validationError && (
              <div className="validation-alert-error">
                {validationError}
              </div>
            )}

            <button 
              className="btn btn-primary action-btn-submit" 
              onClick={handleCreateWorkspace}
              disabled={loading || !targetPath || !labID}
            >
              {loading ? 'Initializing workspace...' : 'Initialize & Open'}
            </button>
          </div>
        )}

        {/* SCREEN 3: Active Workspace Screen (Placeholder) */}
        {screen === 'active-workspace' && (
          <div className="action-form-container glass-card workspace-panel-card">
            <div className="workspace-header-details">
              <span className="status-badge success">
                <CheckIcon /> Active Workspace
              </span>
              <button className="back-btn" onClick={() => setScreen('welcome')}>
                <BackIcon /> Close Workspace
              </button>
            </div>
            
            <h2 className="workspace-path-title">
              {activeWorkspacePath.split('/').pop() || activeWorkspacePath.split('\\').pop()}
            </h2>
            <p className="workspace-full-path">
              📁 {activeWorkspacePath}
            </p>

            <div className="placeholder-details-box">
              <h3>Exercise files loaded.</h3>
              <p>
                The workspace is successfully configured. You can edit your source files directly. 
                Ready to run sandbox tests or submit answers.
              </p>
              <div className="next-steps-list">
                <h4>Next Steps:</h4>
                <ul>
                  <li>Write your code in your preferred text editor.</li>
                  <li>Run <code>euc2 run</code> in your terminal to evaluate public tests.</li>
                </ul>
              </div>
            </div>
          </div>
        )}

      </main>

      {/* VS Code Quick Open Palette Overlay */}
      {showQuickOpen && (
        <div className="quick-open-overlay-blur">
          <div className="quick-open-container" ref={quickOpenRef}>
            <div className="quick-open-input-row">
              <input 
                type="text" 
                value={quickOpenPath} 
                onChange={(e) => setQuickOpenPath(e.target.value)}
                onKeyDown={handleQuickOpenKeyDown}
                placeholder="Search folders..."
                className="quick-open-input"
                autoFocus
              />
              <button className="quick-open-confirm-btn" onClick={handleConfirmQuickOpen}>
                Confirm
              </button>
            </div>

            <div className="quick-open-results-list">
              {quickOpenParent && (
                <div 
                  className="quick-open-item go-up" 
                  onClick={handleGoUp}
                >
                  <span className="item-text">..</span>
                </div>
              )}

              {filteredDirs.length === 0 ? (
                <div className="quick-open-empty-msg">No matching folders found.</div>
              ) : (
                filteredDirs.map((dir, idx) => (
                  <div 
                    key={idx}
                    className="quick-open-item"
                    onClick={() => {
                      const { baseDir, sep } = getPathParts(quickOpenPath);
                      handleQuickOpenNavigate(baseDir + dir + sep);
                    }}
                  >
                    <span className="item-text">{dir}</span>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
