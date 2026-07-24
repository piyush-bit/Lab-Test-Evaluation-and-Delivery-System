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
    activeAdminDrivePath,
    setActiveAdminDrivePath,
    adminDriveManifest,
    adminDriveExercises,
    adminSubmissions,
    handlePrepareAdminSubmission,
    handleAddExerciseToAdminDrive,
    handleDeleteExerciseFromAdminDrive,
    handleGenerateKeyPair,
    quickOpenExercises,
  } = useStudent();

  // Local state for forms
  const [recipientPublicKey, setRecipientPublicKey] = useState('');
  const [generatedPrivateKey, setGeneratedPrivateKey] = useState('');
  const [selectedExId, setSelectedExId] = useState('');
  const [selectedExVer, setSelectedExVer] = useState('');
  const [showAddExModal, setShowAddExModal] = useState(false);
  const [showPrepSubModal, setShowPrepSubModal] = useState(false);

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

  const handleGenerateKey = async () => {
    const keys = await handleGenerateKeyPair();
    if (keys) {
      setRecipientPublicKey(keys.public_key);
      setGeneratedPrivateKey(keys.private_key);
    }
  };

  const formatBytes = (bytes) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  return (
    <div className="app-viewport">
      {/* Background Grid Lines (only when no active drive is open) */}
      {!activeAdminDrivePath && (
        <div className="grid-lines-container">
          <div className="vertical-line v-line-1"></div>
          <div className="vertical-line v-line-2"></div>
          <div className="vertical-line v-line-3"></div>
        </div>
      )}

      {/* Main Viewport Content */}
      <main className="main-viewport-content">
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
          /* ACTIVE DRIVE DETAILS & OPERATIONS VIEW (Same 2-column layout as Dashboard) */
          <div className="vscode-welcome-container">
            <header className="vscode-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <h1>Drive Details & Operations</h1>
                <p style={{ fontFamily: 'var(--font-mono)', fontSize: '0.82rem', marginTop: '4px', opacity: 0.8 }}>
                  {activeAdminDrivePath}
                </p>
              </div>

              <button
                onClick={() => setActiveAdminDrivePath('')}
                style={{
                  padding: '6px 14px',
                  borderRadius: '6px',
                  border: '1px solid var(--border-color)',
                  backgroundColor: 'var(--bg-card)',
                  color: 'var(--text-primary)',
                  fontSize: '0.8rem',
                  cursor: 'pointer',
                  fontWeight: 600
                }}
              >
                ← Close Drive
              </button>
            </header>

            <div className="vscode-columns-grid">
              {/* LEFT COLUMN: EXERCISE INVENTORY */}
              <div className="vscode-left-col">
                <div className="vscode-section">
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '14px' }}>
                    <h2 style={{ fontSize: '0.95rem', display: 'flex', alignItems: 'center', gap: '8px', margin: 0 }}>
                      <PackageIcon size={16} /> Exercise Inventory ({adminDriveExercises.length})
                    </h2>

                    <button
                      onClick={() => setShowAddExModal(!showAddExModal)}
                      style={{
                        padding: '5px 10px',
                        fontSize: '0.75rem',
                        fontWeight: 600,
                        backgroundColor: 'var(--btn-primary-bg)',
                        color: 'var(--btn-primary-text)',
                        border: 'none',
                        borderRadius: '6px',
                        cursor: 'pointer'
                      }}
                    >
                      {showAddExModal ? 'Cancel' : '+ Add Exercise'}
                    </button>
                  </div>

                  {/* Add Exercise Modal / Form */}
                  {showAddExModal && (
                    <div style={{
                      padding: '14px',
                      borderRadius: '8px',
                      backgroundColor: 'var(--bg-card)',
                      border: '1px solid var(--border-color)',
                      marginBottom: '16px'
                    }}>
                      <div style={{ fontSize: '0.78rem', fontWeight: 600, marginBottom: '10px', color: 'var(--text-primary)' }}>Deploy Packaged Exercise to Drive</div>
                      <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                        <input
                          type="text"
                          value={selectedExId}
                          onChange={(e) => setSelectedExId(e.target.value)}
                          placeholder="Exercise ID (e.g. go101-lab01)"
                          style={{
                            padding: '7px 10px',
                            fontSize: '0.8rem',
                            fontFamily: 'var(--font-mono)',
                            backgroundColor: 'var(--bg-terminal)',
                            border: '1px solid var(--border-color)',
                            borderRadius: '6px',
                            color: 'var(--text-primary)'
                          }}
                        />
                        <input
                          type="text"
                          value={selectedExVer}
                          onChange={(e) => setSelectedExVer(e.target.value)}
                          placeholder="Version (e.g. v1.0.0)"
                          style={{
                            padding: '7px 10px',
                            fontSize: '0.8rem',
                            fontFamily: 'var(--font-mono)',
                            backgroundColor: 'var(--bg-terminal)',
                            border: '1px solid var(--border-color)',
                            borderRadius: '6px',
                            color: 'var(--text-primary)'
                          }}
                        />
                        <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end', marginTop: '4px' }}>
                          <button
                            onClick={() => {
                              handleAddExerciseToAdminDrive(activeAdminDrivePath, selectedExId, selectedExVer);
                              setShowAddExModal(false);
                            }}
                            disabled={!selectedExId.trim() || !selectedExVer.trim()}
                            style={{
                              padding: '6px 14px',
                              fontSize: '0.78rem',
                              fontWeight: 600,
                              backgroundColor: 'var(--accent)',
                              color: '#fff',
                              border: 'none',
                              borderRadius: '6px',
                              cursor: selectedExId.trim() && selectedExVer.trim() ? 'pointer' : 'not-allowed'
                            }}
                          >
                            Deploy to Drive
                          </button>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Exercises List Grid */}
                  <div className="vscode-recents-list" style={{ gap: '8px' }}>
                    {adminDriveExercises.length === 0 ? (
                      <div className="empty-recents-msg" style={{ padding: '20px', textAlign: 'center' }}>
                        No exercises deployed on this drive yet.
                      </div>
                    ) : (
                      adminDriveExercises.map((ex, idx) => (
                        <div key={idx} className="vscode-recent-item" style={{
                          display: 'flex',
                          justify: 'space-between',
                          alignItems: 'center',
                          padding: '10px 14px',
                          cursor: 'default'
                        }}>
                          <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                            <span className="recent-icon"><PackageIcon size={16} /></span>
                            <div className="recent-details">
                              <span className="recent-name" style={{ fontSize: '0.88rem', fontWeight: 600 }}>
                                {ex.exercise_id} <span style={{ fontSize: '0.75rem', opacity: 0.6, fontFamily: 'var(--font-mono)' }}>v{ex.version}</span>
                              </span>
                            </div>
                          </div>

                          <button
                            title="Delete exercise from drive"
                            onClick={() => handleDeleteExerciseFromAdminDrive(activeAdminDrivePath, ex.exercise_id, ex.version)}
                            style={{
                              padding: '4px 10px',
                              fontSize: '0.72rem',
                              color: 'var(--accent-red)',
                              backgroundColor: 'rgba(225, 29, 72, 0.08)',
                              border: '1px solid rgba(225, 29, 72, 0.2)',
                              borderRadius: '4px',
                              cursor: 'pointer'
                            }}
                          >
                            Delete
                          </button>
                        </div>
                      ))
                    )}
                  </div>
                </div>
              </div>

              {/* RIGHT COLUMN: SUBMITTED PACKAGES */}
              <div className="vscode-right-col">
                <div className="vscode-section">
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '14px' }}>
                    <h2 style={{ fontSize: '0.95rem', display: 'flex', alignItems: 'center', gap: '8px', margin: 0 }}>
                      <LockIcon size={16} /> Submitted Packages
                    </h2>

                    {adminSubmissions.prepared && (
                      <span style={{
                        fontSize: '0.72rem',
                        padding: '3px 8px',
                        borderRadius: '4px',
                        backgroundColor: 'rgba(5, 150, 105, 0.12)',
                        color: 'var(--accent)',
                        fontWeight: 600
                      }}>
                        🟢 Active
                      </span>
                    )}
                  </div>

                  {!adminSubmissions.prepared ? (
                    /* UNPREPARED SUBMISSION DIRECTORY BANNER */
                    <div style={{
                      padding: '20px',
                      borderRadius: '8px',
                      backgroundColor: 'var(--bg-card)',
                      border: '1px dashed var(--border-color)',
                      textAlign: 'center',
                      display: 'flex',
                      flexDirection: 'column',
                      alignItems: 'center',
                      gap: '12px'
                    }}>
                      <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
                        Submission directory is not prepared on this drive.
                      </div>

                      {!showPrepSubModal ? (
                        <button
                          onClick={() => setShowPrepSubModal(true)}
                          style={{
                            padding: '8px 16px',
                            fontSize: '0.8rem',
                            fontWeight: 600,
                            backgroundColor: 'var(--btn-primary-bg)',
                            color: 'var(--btn-primary-text)',
                            border: 'none',
                            borderRadius: '6px',
                            cursor: 'pointer',
                            display: 'flex',
                            alignItems: 'center',
                            gap: '6px'
                          }}
                        >
                          <ZapIcon size={14} /> Prepare Submission Directory
                        </button>
                      ) : (
                        <div style={{
                          width: '100%',
                          padding: '14px',
                          borderRadius: '8px',
                          backgroundColor: 'var(--bg-main)',
                          border: '1px solid var(--border-color)',
                          textAlign: 'left'
                        }}>
                          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '10px' }}>
                            <span style={{ fontSize: '0.78rem', fontWeight: 600 }}>Recipient Public Key</span>
                            <button
                              onClick={handleGenerateKey}
                              style={{
                                background: 'none',
                                border: 'none',
                                color: 'var(--primary)',
                                fontSize: '0.72rem',
                                cursor: 'pointer',
                                textDecoration: 'underline',
                                display: 'flex',
                                alignItems: 'center',
                                gap: '4px'
                              }}
                            >
                              <ZapIcon size={12} /> Auto-Generate Keypair
                            </button>
                          </div>

                          <input
                            type="text"
                            value={recipientPublicKey}
                            onChange={(e) => setRecipientPublicKey(e.target.value)}
                            placeholder="Base64-encoded X25519 public key..."
                            style={{
                              width: '100%',
                              padding: '8px 12px',
                              fontSize: '0.8rem',
                              fontFamily: 'var(--font-mono)',
                              backgroundColor: 'var(--bg-terminal)',
                              border: '1px solid var(--border-color)',
                              borderRadius: '6px',
                              color: 'var(--text-primary)',
                              marginBottom: '10px'
                            }}
                          />

                          {generatedPrivateKey && (
                            <div style={{
                              padding: '8px 12px',
                              borderRadius: '6px',
                              backgroundColor: 'rgba(234, 179, 8, 0.1)',
                              border: '1px solid rgba(234, 179, 8, 0.2)',
                              color: '#eab308',
                              fontSize: '0.72rem',
                              marginBottom: '10px',
                              wordBreak: 'break-all'
                            }}>
                              <strong>Private Key (Save for grading):</strong> <code>{generatedPrivateKey}</code>
                            </div>
                          )}

                          <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
                            <button
                              onClick={() => setShowPrepSubModal(false)}
                              style={{
                                padding: '6px 12px',
                                fontSize: '0.78rem',
                                backgroundColor: 'transparent',
                                color: 'var(--text-secondary)',
                                border: '1px solid var(--border-color)',
                                borderRadius: '6px',
                                cursor: 'pointer'
                              }}
                            >
                              Cancel
                            </button>
                            <button
                              onClick={() => {
                                handlePrepareAdminSubmission(activeAdminDrivePath, recipientPublicKey);
                                setShowPrepSubModal(false);
                              }}
                              disabled={!recipientPublicKey.trim()}
                              style={{
                                padding: '6px 14px',
                                fontSize: '0.78rem',
                                fontWeight: 600,
                                backgroundColor: 'var(--accent)',
                                color: '#fff',
                                border: 'none',
                                borderRadius: '6px',
                                cursor: recipientPublicKey.trim() ? 'pointer' : 'not-allowed'
                              }}
                            >
                              Confirm
                            </button>
                          </div>
                        </div>
                      )}
                    </div>
                  ) : (
                    /* PREPARED SUBMISSIONS LIST */
                    <div className="vscode-recents-list" style={{ gap: '8px' }}>
                      {(adminSubmissions.submissions || []).length === 0 ? (
                        <div className="empty-recents-msg" style={{ padding: '20px', textAlign: 'center' }}>
                          No student packages submitted yet.
                        </div>
                      ) : (
                        (adminSubmissions.submissions || []).map((sub, idx) => (
                          <div key={idx} className="vscode-recent-item" style={{
                            display: 'flex',
                            justify: 'space-between',
                            alignItems: 'center',
                            padding: '10px 14px',
                            cursor: 'default'
                          }}>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                              <span className="recent-icon"><LockIcon size={16} /></span>
                              <div className="recent-details">
                                <span className="recent-name" style={{ fontSize: '0.85rem', fontWeight: 600, fontFamily: 'var(--font-mono)' }}>
                                  {sub.filename}
                                </span>
                                <span className="recent-path" style={{ fontSize: '0.72rem', opacity: 0.6 }}>
                                  Size: {formatBytes(sub.size)} | Submitted: {new Date(sub.mod_time).toLocaleString()}
                                </span>
                              </div>
                            </div>
                          </div>
                        ))
                      )}
                    </div>
                  )}
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
