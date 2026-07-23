import React, { useState, useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { useStudent } from '../student/StudentContext';
import { HistoryIcon, DriveIcon, PackageIcon, LockIcon, WrenchIcon, ZapIcon, CheckIcon } from '../student/components/Icons';
import SystemStatus from '../student/components/SystemStatus';
import CommandPalette from '../student/components/CommandPalette';

export default function AdminPage() {
  const location = useLocation();
  const {
    validationError,
    setValidationError,
    recentDrives,
    triggerAdminDriveFlow,
    handleOpenAdminDrive,
    handlePrepareAdminDrive,
    activeAdminDrivePath,
    setActiveAdminDrivePath,
    adminDriveManifest,
    adminDriveExercises,
    adminNotification,
    setAdminNotification,
    handlePrepareAdminSubmission,
    handleAddExerciseToAdminDrive,
    quickOpenExercises,
    triggerExercisePicker
  } = useStudent();

  // Local state for active drive forms
  const [recipientPublicKey, setRecipientPublicKey] = useState('');
  const [selectedExId, setSelectedExId] = useState('');
  const [selectedExVer, setSelectedExVer] = useState('');

  // When visiting /admin/drive directly, trigger the admin drive Command Palette flow automatically
  useEffect(() => {
    if (location.pathname === '/admin/drive' || location.pathname === '/admin/drive/') {
      triggerAdminDriveFlow();
    }
  }, [location.pathname]);

  // Sync default values for exercise select
  useEffect(() => {
    if (quickOpenExercises.length > 0 && !selectedExId) {
      setSelectedExId(quickOpenExercises[0].lab_id);
      setSelectedExVer(quickOpenExercises[0].version);
    }
  }, [quickOpenExercises]);

  const generateDemoPublicKey = () => {
    const array = new Uint8Array(32);
    crypto.getRandomValues(array);
    let binary = '';
    for (let i = 0; i < array.length; i++) {
      binary += String.fromCharCode(array[i]);
    }
    const b64Key = btoa(binary);
    setRecipientPublicKey(b64Key);
  };

  return (
    <div className={`app-viewport ${activeAdminDrivePath ? 'workspace-active' : ''}`}>
      {/* Background Grid Lines (only when no active drive is open) */}
      {!activeAdminDrivePath && (
        <div className="grid-lines-container">
          <div className="vertical-line v-line-1"></div>
          <div className="vertical-line v-line-2"></div>
          <div className="vertical-line v-line-3"></div>
        </div>
      )}

      {/* Global Top Bar when active drive is open */}
      {activeAdminDrivePath && (
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
          <div style={{
            fontSize: '0.75rem',
            fontWeight: 700,
            color: 'var(--text-primary)',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
            padding: '4px 8px'
          }}>
            Admin Drive
          </div>

          <div style={{
            fontSize: '0.8rem',
            fontWeight: 600,
            color: 'var(--text-primary)',
            fontFamily: 'var(--font-mono)',
            display: 'flex',
            alignItems: 'center',
            gap: '8px'
          }}>
            <span>{activeAdminDrivePath}</span>
            <span style={{
              fontSize: '0.65rem',
              padding: '2px 6px',
              borderRadius: '4px',
              backgroundColor: 'rgba(5, 150, 105, 0.15)',
              color: 'var(--accent)',
              fontFamily: 'var(--font-sans)',
              fontWeight: 600
            }}>
              🟢 Initialized
            </span>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <button
              title="Close Drive"
              onClick={() => setActiveAdminDrivePath('')}
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

      {/* Main Viewport Content */}
      <main className={`main-viewport-content ${activeAdminDrivePath ? 'workspace-active' : ''}`}>
        {!activeAdminDrivePath ? (
          /* ADMIN DASHBOARD (Matches Student Welcome Dashboard exactly) */
          <div className="vscode-welcome-container">
            <header className="vscode-header">
              <h1>TDES Admin Console</h1>
              <p>Administrative delivery drive and lab evaluation dashboard</p>
              {validationError && (
                <div className="validation-alert-error" style={{ marginTop: '14px' }}>
                  {validationError}
                </div>
              )}
            </header>

            <div className="vscode-columns-grid">
              {/* LEFT COLUMN: Start & Recent Disks */}
              <div className="vscode-left-col">
                <div className="vscode-section">
                  <h2>Start</h2>
                  <div className="vscode-action-list">
                    <div 
                      className="vscode-action-item" 
                      onClick={() => {
                        setValidationError('');
                        triggerAdminDriveFlow();
                      }}
                    >
                      <span className="action-icon"><DriveIcon /></span>
                      <div className="action-details">
                        <span className="action-title">Drive Management</span>
                        <span className="action-desc">Prepare delivery drives, open disks & manage exercises</span>
                      </div>
                    </div>
                  </div>
                </div>

                {/* RECENT DISKS SECTION */}
                <div className="vscode-section" style={{ marginTop: '28px' }}>
                  <h2>Recent Disks</h2>
                  <div className="vscode-recents-list">
                    {recentDrives.length === 0 ? (
                      <div className="empty-recents-msg">No recent disks opened or prepared.</div>
                    ) : (
                      recentDrives.map((diskPath, idx) => (
                        <div 
                          key={idx} 
                          className="vscode-recent-item"
                          onClick={() => handleOpenAdminDrive(diskPath)}
                        >
                          <span className="recent-icon"><HistoryIcon /></span>
                          <div className="recent-details">
                            <span className="recent-name">{diskPath.split('/').pop() || diskPath.split('\\').pop()}</span>
                            <span className="recent-path">{diskPath}</span>
                          </div>
                        </div>
                      ))
                    )}
                  </div>
                </div>
              </div>

              {/* RIGHT COLUMN: System Status */}
              <SystemStatus />
            </div>
          </div>
        ) : (
          /* ACTIVE DRIVE MANAGEMENT VIEW */
          <div className="vscode-welcome-container" style={{ gap: '24px' }}>
            <header className="vscode-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <h1>Drive Details & Operations</h1>
                <p style={{ fontFamily: 'var(--font-mono)', fontSize: '0.85rem' }}>{activeAdminDrivePath}</p>
              </div>

              <button
                onClick={() => setActiveAdminDrivePath('')}
                style={{
                  padding: '6px 14px',
                  borderRadius: '6px',
                  border: '1px solid var(--border-color)',
                  backgroundColor: 'transparent',
                  color: 'var(--text-primary)',
                  fontSize: '0.8rem',
                  cursor: 'pointer'
                }}
              >
                ← Back to Admin Console
              </button>
            </header>

            {adminNotification && (
              <div style={{
                padding: '12px 16px',
                borderRadius: '8px',
                backgroundColor: adminNotification.type === 'success' ? 'rgba(5, 150, 105, 0.1)' : 'rgba(225, 29, 72, 0.1)',
                border: `1px solid ${adminNotification.type === 'success' ? 'rgba(5, 150, 105, 0.3)' : 'rgba(225, 29, 72, 0.3)'}`,
                color: adminNotification.type === 'success' ? 'var(--accent)' : 'var(--accent-red)',
                fontSize: '0.85rem',
                display: 'flex',
                alignItems: 'center',
                gap: '8px'
              }}>
                <span>{adminNotification.type === 'success' ? <CheckIcon size={14} /> : null}</span>
                <span><strong>{adminNotification.title}:</strong> {adminNotification.message}</span>
              </div>
            )}

            <div className="vscode-columns-grid">
              {/* LEFT COLUMN: Stored Exercises on Drive */}
              <div className="vscode-left-col">
                <div className="vscode-section">
                  <h2>Drive Inventory ({adminDriveExercises.length} Exercises)</h2>
                  <div className="vscode-recents-list" style={{ marginTop: '12px' }}>
                    {adminDriveExercises.length === 0 ? (
                      <div className="empty-recents-msg">No exercises deployed on this drive yet.</div>
                    ) : (
                      adminDriveExercises.map((ex, idx) => (
                        <div key={idx} className="vscode-recent-item" style={{ cursor: 'default' }}>
                          <span className="recent-icon"><PackageIcon /></span>
                          <div className="recent-details">
                            <span className="recent-name">{ex.exercise_id} <span className="recent-lab-id">(v{ex.version})</span></span>
                          </div>
                        </div>
                      ))
                    )}
                  </div>
                </div>
              </div>

              {/* RIGHT COLUMN: Drive Operations */}
              <div className="vscode-right-col" style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
                {/* Module 1: Prepare Submission */}
                <div className="vscode-section" style={{
                  padding: '16px',
                  border: '1px solid var(--border-color)',
                  borderRadius: '8px',
                  backgroundColor: 'var(--bg-card)'
                }}>
                  <h2 style={{ fontSize: '0.95rem', marginBottom: '8px', display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <LockIcon size={16} /> Prepare Submission Module
                  </h2>
                  <p style={{ fontSize: '0.78rem', color: 'var(--text-secondary)', marginBottom: '12px' }}>
                    Enable encrypted student submissions on this drive using your X25519 recipient public key.
                  </p>

                  <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <span style={{ fontSize: '0.72rem', fontWeight: 600, color: 'var(--text-muted)' }}>Recipient Public Key (Base64)</span>
                      <button
                        onClick={generateDemoPublicKey}
                        style={{ background: 'none', border: 'none', color: 'var(--primary)', fontSize: '0.72rem', cursor: 'pointer', textDecoration: 'underline', display: 'inline-flex', alignItems: 'center', gap: '4px' }}
                      >
                        <ZapIcon size={12} /> Generate Random Key
                      </button>
                    </div>

                    <input
                      type="text"
                      value={recipientPublicKey}
                      onChange={(e) => setRecipientPublicKey(e.target.value)}
                      placeholder="Base64-encoded X25519 public key..."
                      style={{
                        padding: '8px 12px',
                        fontSize: '0.8rem',
                        fontFamily: 'var(--font-mono)',
                        backgroundColor: 'var(--bg-terminal)',
                        border: '1px solid var(--border-color)',
                        borderRadius: '6px',
                        color: 'var(--text-primary)',
                        outline: 'none'
                      }}
                    />

                    <button
                      onClick={() => handlePrepareAdminSubmission(activeAdminDrivePath, recipientPublicKey)}
                      disabled={!recipientPublicKey.trim()}
                      style={{
                        marginTop: '4px',
                        padding: '8px 14px',
                        fontSize: '0.8rem',
                        fontWeight: 600,
                        backgroundColor: 'var(--btn-primary-bg)',
                        color: 'var(--btn-primary-text)',
                        border: 'none',
                        borderRadius: '6px',
                        cursor: recipientPublicKey.trim() ? 'pointer' : 'not-allowed',
                        opacity: recipientPublicKey.trim() ? 1 : 0.6
                      }}
                    >
                      Enable Drive Submissions
                    </button>
                  </div>
                </div>

                {/* Module 2: Add Exercise */}
                <div className="vscode-section" style={{
                  padding: '16px',
                  border: '1px solid var(--border-color)',
                  borderRadius: '8px',
                  backgroundColor: 'var(--bg-card)'
                }}>
                  <h2 style={{ fontSize: '0.95rem', marginBottom: '8px', display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <PackageIcon size={16} /> Add Exercise to Drive
                  </h2>
                  <p style={{ fontSize: '0.78rem', color: 'var(--text-secondary)', marginBottom: '12px' }}>
                    Deploy a packaged exercise from local cache onto the drive.
                  </p>

                  <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                    <span style={{ fontSize: '0.72rem', fontWeight: 600, color: 'var(--text-muted)' }}>Exercise ID & Version</span>
                    <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: '8px' }}>
                      <input
                        type="text"
                        value={selectedExId}
                        onChange={(e) => setSelectedExId(e.target.value)}
                        placeholder="e.g. go101-lab01"
                        style={{
                          padding: '8px 12px',
                          fontSize: '0.8rem',
                          fontFamily: 'var(--font-mono)',
                          backgroundColor: 'var(--bg-terminal)',
                          border: '1px solid var(--border-color)',
                          borderRadius: '6px',
                          color: 'var(--text-primary)',
                          outline: 'none'
                        }}
                      />
                      <input
                        type="text"
                        value={selectedExVer}
                        onChange={(e) => setSelectedExVer(e.target.value)}
                        placeholder="v1.0"
                        style={{
                          padding: '8px 12px',
                          fontSize: '0.8rem',
                          fontFamily: 'var(--font-mono)',
                          backgroundColor: 'var(--bg-terminal)',
                          border: '1px solid var(--border-color)',
                          borderRadius: '6px',
                          color: 'var(--text-primary)',
                          outline: 'none'
                        }}
                      />
                    </div>

                    <button
                      onClick={() => handleAddExerciseToAdminDrive(activeAdminDrivePath, selectedExId, selectedExVer)}
                      disabled={!selectedExId.trim() || !selectedExVer.trim()}
                      style={{
                        marginTop: '4px',
                        padding: '8px 14px',
                        fontSize: '0.8rem',
                        fontWeight: 600,
                        backgroundColor: 'var(--btn-primary-bg)',
                        color: 'var(--btn-primary-text)',
                        border: 'none',
                        borderRadius: '6px',
                        cursor: selectedExId.trim() && selectedExVer.trim() ? 'pointer' : 'not-allowed',
                        opacity: selectedExId.trim() && selectedExVer.trim() ? 1 : 0.6
                      }}
                    >
                      Deploy Exercise to Drive
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </main>

      <CommandPalette />
    </div>
  );
}
