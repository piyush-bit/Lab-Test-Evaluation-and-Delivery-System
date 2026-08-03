import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useStudent } from '../StudentContext';
import { BackIcon, CheckIcon, ZapIcon } from '../components/Icons';
import Editor from '@monaco-editor/react';
import FileTree from '../components/FileTree';

export default function ActiveWorkspace() {
  const navigate = useNavigate();
  const { 
    activeWorkspacePath, 
    setActiveWorkspacePath, 
    isDarkMode,
    runStatus,
    runResults,
    runOutput,
    showRunPanel,
    setShowRunPanel,
    earnedPoints,
    maxPoints,
    handleRunTests
  } = useStudent();

  const [files, setFiles] = useState([]);
  const [openTabs, setOpenTabs] = useState([]);
  const [selectedTestIdx, setSelectedTestIdx] = useState(0);
  const [copied, setCopied] = useState(false);

  // Compute public tests statistics
  const publicResults = (runResults || []).filter(r => r.public);
  const totalPublic = publicResults.length;
  const passedPublic = publicResults.filter(r => r.status === 'pass').length;
  const publicPassPercentage = totalPublic > 0 ? Math.round((passedPublic / totalPublic) * 100) : 0;

  // Auto-select the first failed test case when run results update
  useEffect(() => {
    if (runResults && runResults.length > 0) {
      const firstFailedIdx = runResults.findIndex(r => r.status !== 'pass');
      setSelectedTestIdx(firstFailedIdx >= 0 ? firstFailedIdx : 0);
    } else {
      setSelectedTestIdx(0);
    }
  }, [runResults]);

  const handleCopyLogs = (text) => {
    if (!text) return;
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };
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
                      <ZapIcon size={14} /> Saving...
                    </span>
                  )}
                  {saveStatus === 'unsaved' && (
                    <span style={{ color: 'var(--accent-red)', fontWeight: 600 }}>
                      ● Unsaved Changes
                    </span>
                  )}
                  {saveStatus === 'saved' && (
                    <span style={{ color: 'var(--accent)', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '4px' }}>
                      Saved
                    </span>
                  )}
                  <span style={{ textTransform: 'uppercase', fontWeight: 600 }}>{getLanguage(activeFilePath)}</span>
                </div>
              </footer>
            </>
          ) : (
            <div style={{
              display: 'flex',
              flex: 1,
              backgroundColor: 'var(--bg-main)',
              alignItems: 'center',
              justifyContent: 'center',
              color: 'var(--text-muted)',
              fontSize: '0.9rem',
              fontWeight: 500,
              userSelect: 'none'
            }}>
              Open any file to edit
            </div>
          )}
        </main>

        {/* Right Side: Run Panel */}
        {showRunPanel && (
          <aside style={{
            width: '380px',
            borderLeft: '1px solid var(--border-color)',
            backgroundColor: 'var(--bg-card)',
            backdropFilter: 'var(--glass-blur)',
            WebkitBackdropFilter: 'var(--glass-blur)',
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden'
          }}>
            {/* Run Panel Header */}
            <div style={{
              padding: '10px 16px',
              borderBottom: '1px solid var(--border-color)',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              backgroundColor: 'rgba(0,0,0,0.02)'
            }}>
              <span style={{ fontSize: '0.72rem', fontWeight: 700, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                Evaluation Results
              </span>
              <button 
                onClick={() => setShowRunPanel(false)}
                title="Hide Panel" 
                style={{
                  background: 'none',
                  border: 'none',
                  cursor: 'pointer',
                  color: 'var(--text-muted)',
                  fontSize: '1rem',
                  display: 'flex',
                  alignItems: 'center',
                  padding: '2px'
                }}
              >
                ×
              </button>
            </div>

            {/* Run Panel Contents */}
            <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflowY: 'auto' }}>
              {/* Overall Score Progress Card */}
              {runStatus !== 'idle' && (
                <div style={{ padding: '16px', borderBottom: '1px solid var(--border-color)', backgroundColor: 'rgba(0,0,0,0.01)' }}>
                  {runStatus === 'running' ? (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                      <div style={{ fontSize: '0.8rem', fontWeight: 600, color: 'var(--text-secondary)', display: 'flex', alignItems: 'center', gap: '4px' }}>
                        <ZapIcon size={14} /> Executing test cases...
                      </div>
                      <div style={{ height: '6px', width: '100%', backgroundColor: 'var(--border-color)', borderRadius: '4px', overflow: 'hidden', position: 'relative' }}>
                        <div style={{ height: '100%', width: '40%', backgroundColor: 'var(--primary)', borderRadius: '4px', position: 'absolute', animation: 'indeterminate-progress 1.5s infinite ease-in-out' }}></div>
                      </div>
                    </div>
                  ) : (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
                        <span style={{ fontSize: '0.8rem', fontWeight: 700, color: 'var(--text-primary)' }}>
                          Public Tests
                        </span>
                        <span style={{ fontSize: '1.1rem', fontWeight: 800, color: totalPublic > 0 && passedPublic === totalPublic ? 'var(--accent)' : 'var(--text-primary)' }}>
                          {passedPublic} <span style={{ fontSize: '0.75rem', fontWeight: 500, color: 'var(--text-muted)' }}>/ {totalPublic} passed</span>
                        </span>
                      </div>
                      {/* Premium Progress Bar */}
                      <div style={{ height: '6px', width: '100%', backgroundColor: 'var(--border-color)', borderRadius: '4px', overflow: 'hidden' }}>
                        <div style={{
                          height: '100%',
                          width: `${publicPassPercentage}%`,
                          backgroundColor: totalPublic > 0 && passedPublic === totalPublic ? 'var(--accent)' : passedPublic > 0 ? '#eab308' : '#f43f5e',
                          borderRadius: '4px',
                          transition: 'width 0.4s ease, background-color 0.4s ease'
                        }}></div>
                      </div>
                      <div style={{ fontSize: '0.68rem', color: 'var(--text-muted)', textAlign: 'right' }}>
                        {totalPublic > 0 ? `${publicPassPercentage}% Completed` : '0%'}
                      </div>
                    </div>
                  )}
                </div>
              )}

              {/* Individual Test Cases List */}
              <div style={{ flex: '0 0 auto', borderBottom: '1px solid var(--border-color)' }}>
                <div style={{ padding: '8px 16px', fontSize: '0.68rem', fontWeight: 700, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.05em', backgroundColor: 'rgba(0,0,0,0.01)', borderBottom: '1px solid var(--border-color)' }}>
                  Test Cases ({runResults.length})
                </div>
                {runResults.length > 0 ? (
                  <div style={{ display: 'flex', flexDirection: 'column' }}>
                    {runResults.map((tr, idx) => {
                      const isActive = selectedTestIdx === idx;
                      const isPassed = tr.status === 'pass';
                      return (
                        <div 
                          key={idx}
                          onClick={() => setSelectedTestIdx(idx)}
                          style={{
                            display: 'flex',
                            justifyContent: 'space-between',
                            alignItems: 'center',
                            padding: '10px 16px',
                            borderBottom: idx < runResults.length - 1 ? '1px solid var(--border-color)' : 'none',
                            fontSize: '0.78rem',
                            cursor: 'pointer',
                            backgroundColor: isActive ? 'rgba(79, 70, 229, 0.04)' : 'transparent',
                            borderLeft: isActive ? '3px solid var(--primary)' : '3px solid transparent',
                            transition: 'all 0.15s ease'
                          }}
                          className="test-case-row-item"
                          onMouseEnter={(e) => {
                            if (!isActive) e.currentTarget.style.backgroundColor = 'rgba(0,0,0,0.02)';
                            const playBtn = e.currentTarget.querySelector('.run-single-btn');
                            if (playBtn) playBtn.style.opacity = '1';
                          }}
                          onMouseLeave={(e) => {
                            if (!isActive) e.currentTarget.style.backgroundColor = 'transparent';
                            const playBtn = e.currentTarget.querySelector('.run-single-btn');
                            if (playBtn) playBtn.style.opacity = '0';
                          }}
                        >
                          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', overflow: 'hidden', marginRight: '8px' }}>
                            <span style={{ display: 'flex', alignItems: 'center', flexShrink: 0 }}>
                              {!tr.public ? (
                                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="var(--text-muted)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{ opacity: 0.7 }}>
                                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                                  <path d="M7 11V7a5 5 0 0 1 10 0v4" />
                                </svg>
                              ) : tr.status === 'running' ? (
                                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="var(--primary)" strokeWidth="3.5" strokeLinecap="round" style={{ animation: 'spin 1s linear infinite' }}>
                                  <circle cx="12" cy="12" r="10" strokeDasharray="30" strokeDashoffset="10" />
                                </svg>
                              ) : tr.status === 'pass' ? (
                                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="var(--accent)" strokeWidth="3.5" strokeLinecap="round" strokeLinejoin="round">
                                  <polyline points="20 6 9 17 4 12" />
                                </svg>
                              ) : tr.status === 'fail' ? (
                                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#f43f5e" strokeWidth="3.5" strokeLinecap="round" strokeLinejoin="round">
                                  <line x1="18" y1="6" x2="6" y2="18" />
                                  <line x1="6" y1="6" x2="18" y2="18" />
                                </svg>
                              ) : (
                                <div style={{ width: '8px', height: '8px', borderRadius: '50%', border: '2px solid var(--text-muted)', margin: '2px' }}></div>
                              )}
                            </span>
                            <span style={{ 
                              fontFamily: 'var(--font-mono)', 
                              fontSize: 11, 
                              color: !tr.public ? 'var(--text-muted)' : isActive ? 'var(--text-primary)' : 'var(--text-secondary)', 
                              overflow: 'hidden', 
                              textOverflow: 'ellipsis', 
                              whiteSpace: 'nowrap' 
                            }}>
                              {tr.command}
                            </span>
                            {/* Visibility Badge */}
                            {tr.public ? (
                              <span style={{
                                fontSize: '0.55rem',
                                fontWeight: 700,
                                color: 'var(--accent)',
                                backgroundColor: 'rgba(16, 185, 129, 0.08)',
                                border: '1px solid rgba(16, 185, 129, 0.15)',
                                padding: '0px 4px',
                                borderRadius: '4px',
                                textTransform: 'uppercase',
                                letterSpacing: '0.02em',
                                flexShrink: 0
                              }}>
                                Public
                              </span>
                            ) : (
                              <span style={{
                                fontSize: '0.55rem',
                                fontWeight: 700,
                                color: 'var(--text-muted)',
                                backgroundColor: 'rgba(0,0,0,0.02)',
                                border: '1px solid var(--border-color)',
                                padding: '0px 4px',
                                borderRadius: '4px',
                                textTransform: 'uppercase',
                                letterSpacing: '0.02em',
                                flexShrink: 0
                              }}>
                                Private
                              </span>
                            )}
                          </div>
                          
                          {/* Right Side: Run button / Lock symbol */}
                          {tr.public ? (
                            <button
                              className="run-single-btn"
                              onClick={(e) => {
                                e.stopPropagation();
                                handleRunTests(tr.command);
                              }}
                              disabled={runStatus === 'running' || tr.status === 'running'}
                              title="Run this test case"
                              style={{
                                background: 'none',
                                border: 'none',
                                cursor: 'pointer',
                                padding: '2px 6px',
                                color: 'var(--text-muted)',
                                display: 'inline-flex',
                                alignItems: 'center',
                                borderRadius: '4px',
                                opacity: 0,
                                transition: 'opacity 0.15s ease, color 0.15s ease'
                              }}
                              onMouseEnter={(e) => e.currentTarget.style.color = 'var(--primary)'}
                              onMouseLeave={(e) => e.currentTarget.style.color = 'var(--text-muted)'}
                            >
                              <svg width="10" height="10" viewBox="0 0 24 24" fill="currentColor">
                                <polygon points="5 3 19 12 5 21" />
                              </svg>
                            </button>
                          ) : (
                            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="var(--text-muted)" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{ opacity: 0.4, marginRight: '4px' }}>
                              <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                              <path d="M7 11V7a5 5 0 0 1 10 0v4" />
                            </svg>
                          )}
                        </div>
                      );
                    })}
                  </div>
                ) : (
                  <div style={{ padding: '24px 16px', fontSize: '0.75rem', color: 'var(--text-muted)', fontStyle: 'italic', textAlign: 'center' }}>
                    {runStatus === 'running' ? 'Running test commands...' : runStatus === 'idle' ? 'No tests executed yet' : 'No test cases declared in manifest'}
                  </div>
                )}
              </div>

              {/* Console Logs Section */}
              <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minHeight: '220px' }}>
                <div style={{
                  padding: '6px 16px',
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  backgroundColor: 'rgba(0,0,0,0.02)',
                  borderBottom: '1px solid var(--border-color)',
                  borderTop: '1px solid var(--border-color)'
                }}>
                  <span style={{ fontSize: '0.68rem', fontWeight: 700, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                    {runResults.length > 0 && runResults[selectedTestIdx]
                      ? `Output: ${runResults[selectedTestIdx].command}`
                      : 'Console Output Log'
                    }
                  </span>
                  {(runOutput || (runResults.length > 0 && runResults[selectedTestIdx]?.output)) && (
                    <button
                      onClick={() => handleCopyLogs(
                        runResults.length > 0 && runResults[selectedTestIdx]
                          ? runResults[selectedTestIdx].output
                          : runOutput
                      )}
                      style={{
                        background: 'none',
                        border: '1px solid var(--border-color)',
                        backgroundColor: 'var(--bg-main)',
                        borderRadius: '4px',
                        padding: '2px 8px',
                        fontSize: '0.65rem',
                        fontWeight: 600,
                        color: 'var(--text-secondary)',
                        cursor: 'pointer'
                      }}
                    >
                      {copied ? 'Copied!' : 'Copy'}
                    </button>
                  )}
                </div>
                <pre style={{
                  flex: 1,
                  margin: 0,
                  padding: '12px 16px',
                  backgroundColor: 'var(--bg-terminal)',
                  color: '#e6edf3',
                  fontFamily: 'var(--font-mono)',
                  fontSize: '0.7rem',
                  lineHeight: '1.45',
                  overflow: 'auto',
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-all',
                  textAlign: 'left'
                }}>
                  {runResults.length > 0 && runResults[selectedTestIdx]
                    ? runResults[selectedTestIdx].output || 'Command executed with no output.'
                    : runOutput || 'No output log.'
                  }
                </pre>
              </div>
            </div>
          </aside>
        )}
      </div>
    </div>
  );
}
