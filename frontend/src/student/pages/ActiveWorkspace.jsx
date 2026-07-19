import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useStudent } from '../StudentContext';
import { BackIcon, CheckIcon } from '../components/Icons';
import Editor from '@monaco-editor/react';
import FileTree from '../components/FileTree';

export default function ActiveWorkspace() {
  const navigate = useNavigate();
  const { activeWorkspacePath, setActiveWorkspacePath, isDarkMode } = useStudent();

  const [files, setFiles] = useState([]);
  const [openTabs, setOpenTabs] = useState([]);
  const [activeFilePath, setActiveFilePath] = useState(null);
  const [fileContent, setFileContent] = useState('');
  const [saveStatus, setSaveStatus] = useState('saved'); // 'saved', 'saving', 'unsaved'
  const [loadingContent, setLoadingContent] = useState(false);
  const lastLoadedContentRef = useRef('');
  const debounceTimerRef = useRef(null);

  // If there's no active workspace, redirect back to welcome
  useEffect(() => {
    if (!activeWorkspacePath) {
      navigate('/');
    }
  }, [activeWorkspacePath, navigate]);

  const fetchFiles = async () => {
    try {
      const res = await fetch(`/api/workspace/files?path=${encodeURIComponent(activeWorkspacePath)}`);
      if (res.ok) {
        const data = await res.json();
        setFiles(data);
      }
    } catch (err) {
      console.error("Error fetching files:", err);
    }
  };

  useEffect(() => {
    if (activeWorkspacePath) {
      fetchFiles();
    }
  }, [activeWorkspacePath]);

  const handleFileClick = async (node) => {
    if (!openTabs.includes(node.path)) {
      setOpenTabs([...openTabs, node.path]);
    }
    setActiveFilePath(node.path);
  };

  const handleCloseTab = (pathToDelete, e) => {
    e.stopPropagation();
    const filtered = openTabs.filter(p => p !== pathToDelete);
    setOpenTabs(filtered);
    if (activeFilePath === pathToDelete) {
      if (filtered.length > 0) {
        setActiveFilePath(filtered[filtered.length - 1]);
      } else {
        setActiveFilePath(null);
        setFileContent('');
      }
    }
  };

  useEffect(() => {
    if (!activeFilePath) return;

    const fetchContent = async () => {
      setLoadingContent(true);
      try {
        const res = await fetch(`/api/workspace/file?path=${encodeURIComponent(activeFilePath)}`);
        if (res.ok) {
          const data = await res.json();
          setFileContent(data.content);
          lastLoadedContentRef.current = data.content;
          setSaveStatus('saved');
        }
      } catch (err) {
        console.error("Error fetching file content:", err);
      }
      setLoadingContent(false);
    };

    fetchContent();
  }, [activeFilePath]);

  const handleEditorChange = (value) => {
    setFileContent(value);
    if (value !== lastLoadedContentRef.current) {
      setSaveStatus('unsaved');
    }
  };

  // Autosave effect with 1-second debounce
  useEffect(() => {
    if (!activeFilePath || saveStatus !== 'unsaved') return;

    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
    }

    debounceTimerRef.current = setTimeout(async () => {
      setSaveStatus('saving');
      try {
        const res = await fetch('/api/workspace/file', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            path: activeFilePath,
            content: fileContent
          })
        });

        if (res.ok) {
          lastLoadedContentRef.current = fileContent;
          setSaveStatus('saved');
        } else {
          setSaveStatus('unsaved');
        }
      } catch (err) {
        console.error("Error autosaving file:", err);
        setSaveStatus('unsaved');
      }
    }, 1000);

    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
      }
    };
  }, [fileContent, activeFilePath, saveStatus]);

  const getLanguage = (filePath) => {
    if (!filePath) return 'plaintext';
    const ext = filePath.split('.').pop().toLowerCase();
    switch (ext) {
      case 'go': return 'go';
      case 'py': return 'python';
      case 'js': case 'jsx': return 'javascript';
      case 'ts': case 'tsx': return 'typescript';
      case 'json': return 'json';
      case 'md': return 'markdown';
      case 'html': return 'html';
      case 'css': return 'css';
      case 'sh': return 'shell';
      case 'yaml': case 'yml': return 'yaml';
      default: return 'plaintext';
    }
  };

  if (!activeWorkspacePath) return null;

  return (
    <div className="active-workspace-panel glass-card" style={{ display: 'flex', flexDirection: 'column', height: '100%', margin: 0, padding: 0, overflow: 'hidden', border: 'none', borderRadius: 0 }}>

      {/* Main Split View: Sidebar + Editor */}
      <div style={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
        {/* Left Side: Sidebar Explorer */}
        <aside style={{ width: '260px', borderRight: '1px solid var(--border-color)', backgroundColor: 'var(--bg-card)', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
          <div style={{ padding: '10px 16px', borderBottom: '1px solid var(--border-color)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span style={{ fontSize: '0.72rem', fontWeight: 700, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              Files Explorer
            </span>
            <button 
              onClick={fetchFiles}
              title="Refresh Files" 
              style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-secondary)', padding: '2px', display: 'flex', alignItems: 'center' }}
            >
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21.5 2v6h-6M21.34 15.57a10 10 0 1 1-.57-8.38l5.67-5.67"/>
              </svg>
            </button>
          </div>
          <div style={{ flex: 1, overflowY: 'auto', padding: '10px' }}>
            <FileTree 
              files={files} 
              onFileClick={handleFileClick} 
              activeFilePath={activeFilePath} 
            />
          </div>
        </aside>

        {/* Right Side: Workspace Area */}
        <main style={{ flex: 1, display: 'flex', flexDirection: 'column', backgroundColor: 'var(--bg-main)', overflow: 'hidden' }}>
          {openTabs.length > 0 ? (
            <>
              {/* Tab Bar */}
              <div className="tab-bar-container" style={{ display: 'flex', borderBottom: '1px solid var(--border-color)', backgroundColor: 'var(--bg-card)', overflowX: 'auto', height: '35px' }}>
                {openTabs.map((tabPath) => {
                  const filename = tabPath.split('/').pop().split('\\').pop();
                  const isActive = activeFilePath === tabPath;
                  return (
                    <div 
                      key={tabPath}
                      onClick={() => setActiveFilePath(tabPath)}
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '8px',
                        padding: '0 16px',
                        height: '100%',
                        cursor: 'pointer',
                        fontSize: '0.78rem',
                        fontWeight: isActive ? 600 : 400,
                        color: isActive ? 'var(--primary)' : 'var(--text-secondary)',
                        backgroundColor: isActive ? 'var(--bg-main)' : 'rgba(0,0,0,0.02)',
                        borderRight: '1px solid var(--border-color)',
                        borderTop: isActive ? '2px solid var(--primary)' : 'none',
                        transition: 'all 0.15s ease',
                        position: 'relative'
                      }}
                    >
                      <span>{filename}</span>
                      <button 
                        onClick={(e) => handleCloseTab(tabPath, e)}
                        style={{
                          background: 'none',
                          border: 'none',
                          cursor: 'pointer',
                          color: 'var(--text-muted)',
                          fontSize: '0.75rem',
                          display: 'flex',
                          alignItems: 'center',
                          padding: '2px',
                          borderRadius: '4px'
                        }}
                        onMouseEnter={(e) => e.target.style.color = 'var(--accent-red)'}
                        onMouseLeave={(e) => e.target.style.color = 'var(--text-muted)'}
                      >
                        ×
                      </button>
                    </div>
                  );
                })}
              </div>

              {/* Editor Workspace */}
              <div style={{ flex: 1, position: 'relative', overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
                {loadingContent ? (
                  <div style={{ display: 'flex', flex: 1, alignItems: 'center', justifyContent: 'center', color: 'var(--text-secondary)' }}>
                    Loading file content...
                  </div>
                ) : (
                  <Editor
                    height="100%"
                    language={getLanguage(activeFilePath)}
                    value={fileContent}
                    onChange={handleEditorChange}
                    theme={isDarkMode ? 'vs-dark' : 'light'}
                    options={{
                      fontSize: 14,
                      minimap: { enabled: true },
                      scrollBeyondLastLine: false,
                      automaticLayout: true,
                      tabSize: 4,
                      wordWrap: 'on'
                    }}
                  />
                )}
              </div>

              {/* Status Bar */}
              <footer style={{ height: '24px', borderTop: '1px solid var(--border-color)', backgroundColor: 'var(--bg-card)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '0 16px', fontSize: '0.72rem', color: 'var(--text-secondary)' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <span style={{ fontWeight: 500 }}>Active file:</span>
                  <span style={{ fontFamily: 'var(--font-mono)' }}>{activeFilePath.split('/').pop().split('\\').pop()}</span>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                  {saveStatus === 'saving' && (
                    <span style={{ color: 'var(--secondary)', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '4px' }}>
                      ⚡ Saving...
                    </span>
                  )}
                  {saveStatus === 'unsaved' && (
                    <span style={{ color: 'var(--accent-red)', fontWeight: 600 }}>
                      ● Unsaved Changes
                    </span>
                  )}
                  {saveStatus === 'saved' && (
                    <span style={{ color: 'var(--accent)', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '4px' }}>
                      ✓ Saved
                    </span>
                  )}
                  <span style={{ textTransform: 'uppercase', fontWeight: 600 }}>{getLanguage(activeFilePath)}</span>
                </div>
              </footer>
            </>
          ) : (
            <div style={{ display: 'flex', flex: 1, backgroundColor: 'var(--bg-main)' }}></div>
          )}
        </main>
      </div>
    </div>
  );
}
