import React, { useState, useEffect } from 'react';
import { useStudent } from '../student/StudentContext';
import { WrenchIcon, ClipboardIcon } from '../student/components/Icons';
import StudentManager from './StudentManager';
import EvaluationsManager from './EvaluationsManager';

export default function ServerManager({ onBack }) {
  const { activeRegistryServer } = useStudent();
  const [activeSubView, setActiveSubView] = useState('main'); // 'main' | 'students' | 'evaluations'

  const serverURL = activeRegistryServer?.url || 'http://localhost:8080';
  const bearerToken = activeRegistryServer?.token || '';

  const [serverOnline, setServerOnline] = useState(true);
  const [checkingHealth, setCheckingHealth] = useState(false);
  const [healthError, setHealthError] = useState('');

  const checkHealth = async () => {
    if (!serverURL) return;
    setCheckingHealth(true);
    setHealthError('');
    try {
      const resp = await fetch(`/api/remote/health?url=${encodeURIComponent(serverURL)}`);
      const data = await resp.json();
      if (data.online) {
        setServerOnline(true);
      } else {
        setServerOnline(false);
        setHealthError(data.error || 'Server unreachable');
      }
    } catch (err) {
      setServerOnline(false);
      setHealthError(err.message);
    } finally {
      setCheckingHealth(false);
    }
  };

  useEffect(() => {
    checkHealth();
    const interval = setInterval(checkHealth, 5000);
    return () => clearInterval(interval);
  }, [serverURL, bearerToken]);

  if (activeSubView === 'students') {
    return <StudentManager onBack={() => setActiveSubView('main')} />;
  }

  if (activeSubView === 'evaluations') {
    return <EvaluationsManager onBack={() => setActiveSubView('main')} />;
  }

  return (
    <div className="vscode-welcome-container" style={{ textAlign: 'left', maxWidth: '1100px', margin: '0 auto' }}>
      {/* SINGLE HEADER */}
      <header className="vscode-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1>Registry Management</h1>
          <p style={{ fontFamily: 'var(--font-mono)', fontSize: '0.82rem', marginTop: '4px', opacity: 0.8 }}>
            {serverURL}
          </p>
        </div>
        <button
          onClick={onBack}
          style={{ background: 'none', border: 'none', color: 'var(--text-secondary)', fontSize: '0.8rem', cursor: 'pointer', fontWeight: 600 }}
        >
          ← Back to Dashboard
        </button>
      </header>

      {/* 2-COLUMN GRID SYSTEM */}
      <div className="vscode-columns-grid" style={{ marginTop: '28px' }}>
        {/* LEFT COLUMN: Start Options */}
        <div className="vscode-left-col">
          <div className="vscode-section">
            <h2>Start</h2>
            <div className="vscode-action-list">
              <div
                className="vscode-action-item"
                onClick={() => setActiveSubView('students')}
              >
                <span className="recent-icon"><WrenchIcon /></span>
                <div className="action-details">
                  <span className="action-title">Student Management</span>
                  <span className="action-desc">Manage student roster, TOFU PIN activation state & onboarding</span>
                </div>
              </div>

              <div
                className="vscode-action-item"
                onClick={() => setActiveSubView('evaluations')}
              >
                <span className="recent-icon"><ClipboardIcon /></span>
                <div className="action-details">
                  <span className="action-title">Evaluations</span>
                  <span className="action-desc">View, search, filter, and inspect recent lab submission evaluation records</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* RIGHT COLUMN: Server Status */}
        <div className="vscode-right-col">
          <div className="vscode-section">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
              <h2 style={{ margin: 0 }}>Server Status</h2>
              <button
                onClick={checkHealth}
                disabled={checkingHealth}
                title="Refresh Connection Status"
                style={{
                  background: 'none',
                  border: 'none',
                  color: 'var(--text-secondary)',
                  fontSize: '0.95rem',
                  cursor: 'pointer',
                  opacity: checkingHealth ? 0.5 : 1,
                  display: 'inline-flex',
                  alignItems: 'center',
                  padding: 0
                }}
              >
                ↻
              </button>
            </div>

            <div style={{ fontSize: '0.8rem', display: 'flex', flexDirection: 'column', gap: '12px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid var(--border-color)', paddingBottom: '10px' }}>
                <span style={{ color: 'var(--text-secondary)' }}>Connection Status:</span>
                <span style={{
                  fontSize: '0.72rem',
                  fontWeight: 700,
                  padding: '2px 8px',
                  borderRadius: '4px',
                  backgroundColor: serverOnline ? 'rgba(16, 185, 129, 0.12)' : 'rgba(239, 68, 68, 0.12)',
                  color: serverOnline ? 'var(--accent)' : 'var(--accent-red)',
                  textTransform: 'uppercase'
                }}>
                  {serverOnline ? '● Online' : '● Offline'}
                </span>
              </div>

              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid var(--border-color)', paddingBottom: '10px' }}>
                <span style={{ color: 'var(--text-secondary)' }}>Bearer Token:</span>
                <span style={{
                  fontSize: '0.72rem',
                  fontWeight: 600,
                  padding: '2px 6px',
                  borderRadius: '4px',
                  backgroundColor: bearerToken ? 'rgba(79, 70, 229, 0.1)' : 'rgba(255, 255, 255, 0.05)',
                  color: bearerToken ? 'var(--primary)' : 'var(--text-muted)'
                }}>
                  {bearerToken ? 'Authenticated Token Set' : 'Unauthenticated'}
                </span>
              </div>
            </div>

            {healthError && (
              <div className="validation-alert-error" style={{ marginTop: '16px', textAlign: 'left' }}>
                Server Unreachable: {healthError}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
