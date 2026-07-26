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
    inspectAdminDrive,
    handlePrepareAdminSubmission,
    handleAddExerciseToAdminDrive,
    triggerAdminAddExercisePicker,
    triggerAdminConfirmClearResults,
    triggerAdminConfirmClearSubmissions,
    handleDeleteExerciseFromAdminDrive,
    handleGenerateKeyPair,
    handleBatchEvaluateSubmissions,
    handleSingleEvaluateSubmission,
    handleSavePrivateKeyRecord,
    quickOpenExercises,
  } = useStudent();

  // Local state for forms & modals
  const [recipientPublicKey, setRecipientPublicKey] = useState('');
  const [generatedPrivateKey, setGeneratedPrivateKey] = useState('');
  const [instructorPrivateKey, setInstructorPrivateKey] = useState('');
  const [showPrepSubModal, setShowPrepSubModal] = useState(false);
  const [showEvalKeyModal, setShowEvalKeyModal] = useState(false);
  const [evalTargetFilename, setEvalTargetFilename] = useState('');
  const [evaluatingTarget, setEvaluatingTarget] = useState(null); // null, '__batch__', or filename
  const [evalResults, setEvalResults] = useState(null);
  const [expandedLogs, setExpandedLogs] = useState({});

  const toggleLogExpand = (idx) => {
    setExpandedLogs(prev => ({ ...prev, [idx]: !prev[idx] }));
  };

  // When visiting /admin/drive directly, trigger the admin drive Command Palette flow automatically
  useEffect(() => {
    if (location.pathname === '/admin/drive' || location.pathname === '/admin/drive/') {
      triggerAdminDriveFlow();
    }
  }, [location.pathname]);

  const handleGenerateKey = async () => {
    const keys = await handleGenerateKeyPair();
    if (keys) {
      setRecipientPublicKey(keys.public_key);
      setGeneratedPrivateKey(keys.private_key);
      setInstructorPrivateKey(keys.private_key);
    }
  };

  const formatBytes = (bytes) => {
    if (!bytes || bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const executeEvaluationDirect = async (filename, keyToUse = '') => {
    const targetKey = filename || '__batch__';
    setEvaluatingTarget(targetKey);
    const key = keyToUse || adminSubmissions.private_key || instructorPrivateKey;

    if (filename) {
      const record = await handleSingleEvaluateSubmission(activeAdminDrivePath, filename, key);
      if (record) {
        setEvalResults([record]);
      }
    } else {
      const records = await handleBatchEvaluateSubmissions(activeAdminDrivePath, key);
      if (records) {
        setEvalResults(records);
      }
    }
    setEvaluatingTarget(null);
    setShowEvalKeyModal(false);
  };

  const startBatchEvaluation = () => {
    if (adminSubmissions.has_private_key) {
      executeEvaluationDirect('');
    } else {
      setEvalTargetFilename('');
      setShowEvalKeyModal(true);
    }
  };

  const startSingleEvaluation = (filename) => {
    if (adminSubmissions.has_private_key) {
      executeEvaluationDirect(filename);
    } else {
      setEvalTargetFilename(filename);
      setShowEvalKeyModal(true);
    }
  };

  const handleSaveMissingKey = async () => {
    if (!instructorPrivateKey.trim()) return;
    const saved = await handleSavePrivateKeyRecord(adminSubmissions.submission_id, adminSubmissions.public_key, instructorPrivateKey.trim());
    if (saved) {
      await inspectAdminDrive(activeAdminDrivePath);
      executeEvaluationDirect(evalTargetFilename, instructorPrivateKey.trim());
    }
  };

  const exportCSVReport = () => {
    if (!evalResults || evalResults.length === 0) return;
    const headers = ['Student ID', 'Lab ID', 'Version', 'Status', 'Earned Points', 'Max Points', 'Error'];
    const rows = evalResults.map(r => [
      r.student_id || 'Unknown',
      r.lab_id || 'N/A',
      r.version || 'N/A',
      r.status || 'error',
      r.earned_points || 0,
      r.max_points || 0,
      `"${(r.error || '').replace(/"/g, '""')}"`
    ]);
    const csvContent = 'data:text/csv;charset=utf-8,' + [headers.join(','), ...rows.map(e => e.join(','))].join('\n');
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement('a');
    link.setAttribute('href', encodedUri);
    link.setAttribute('download', `evaluation_report_${Date.now()}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
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
          /* ACTIVE DRIVE DETAILS VIEW (Minimal VS Code Dashboard Design Language) */
          <div className="vscode-welcome-container">
            <header className="vscode-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <h1>Drive Details & Operations</h1>
                <p style={{ fontFamily: 'var(--font-mono)', fontSize: '0.82rem', marginTop: '4px', opacity: 0.8 }}>
                  {activeAdminDrivePath}
                </p>
              </div>

              <button
                onClick={() => {
                  setActiveAdminDrivePath('');
                  setEvalResults(null);
                }}
                style={{
                  background: 'none',
                  border: 'none',
                  color: 'var(--text-secondary)',
                  fontSize: '0.8rem',
                  cursor: 'pointer',
                  fontWeight: 600
                }}
              >
                ← Close Drive
              </button>
            </header>

            {validationError && (
              <div className="validation-alert-error" style={{ marginBottom: '14px' }}>
                {validationError}
              </div>
            )}

            <div className="vscode-columns-grid">
              {/* LEFT COLUMN: EXERCISE INVENTORY */}
              <div className="vscode-left-col">
                <div className="vscode-section">
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px', borderBottom: '1px solid var(--border-color)', paddingBottom: '6px' }}>
                    <h2 style={{ border: 'none', padding: 0, margin: 0 }}>
                      Exercise Inventory ({adminDriveExercises.length})
                    </h2>

                    <button
                      onClick={() => triggerAdminAddExercisePicker(activeAdminDrivePath)}
                      style={{
                        background: 'none',
                        border: 'none',
                        color: 'var(--primary)',
                        fontSize: '0.78rem',
                        fontWeight: 600,
                        cursor: 'pointer',
                        padding: 0
                      }}
                    >
                      + Add Exercise
                    </button>
                  </div>

                  <div className="vscode-recents-list">
                    {adminDriveExercises.length === 0 ? (
                      <div className="empty-recents-msg">No exercises deployed on this drive.</div>
                    ) : (
                      adminDriveExercises.map((ex, idx) => (
                        <div key={idx} className="vscode-recent-item" style={{ justifyContent: 'space-between' }}>
                          <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                            <span className="recent-icon"><PackageIcon size={14} /></span>
                            <div className="recent-details">
                              <span className="recent-name">
                                {ex.exercise_id} <span style={{ fontSize: '0.72rem', color: 'var(--text-muted)', fontFamily: 'var(--font-mono)' }}>v{ex.version}</span>
                              </span>
                            </div>
                          </div>

                          <button
                            title="Delete exercise"
                            onClick={() => handleDeleteExerciseFromAdminDrive(activeAdminDrivePath, ex.exercise_id, ex.version)}
                            style={{
                              background: 'none',
                              border: 'none',
                              color: 'var(--accent-red)',
                              fontSize: '0.75rem',
                              cursor: 'pointer',
                              opacity: 0.8
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
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px', borderBottom: '1px solid var(--border-color)', paddingBottom: '6px' }}>
                    <h2 style={{ border: 'none', padding: 0, margin: 0 }}>
                      Submitted Packages ({adminSubmissions.submissions?.length || 0})
                    </h2>

                    {adminSubmissions.prepared && (adminSubmissions.submissions?.length || 0) > 0 && (
                      <button
                        onClick={startBatchEvaluation}
                        disabled={evaluatingTarget !== null}
                        style={{
                          background: 'none',
                          border: 'none',
                          color: evaluatingTarget !== null ? 'var(--text-muted)' : 'var(--primary)',
                          fontSize: '0.78rem',
                          fontWeight: 600,
                          cursor: evaluatingTarget !== null ? 'not-allowed' : 'pointer',
                          padding: 0
                        }}
                      >
                        {evaluatingTarget === '__batch__' ? 'Evaluating All...' : 'Batch Evaluate All'}
                      </button>
                    )}
                  </div>

                  {/* Submission ID Minimal Status Bar */}
                  {adminSubmissions.prepared && (
                    <div style={{
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                      fontSize: '0.74rem',
                      fontFamily: 'var(--font-mono)',
                      color: 'var(--text-muted)',
                      marginTop: '6px',
                      marginBottom: '16px'
                    }}>
                      <span>ID: {adminSubmissions.submission_id || 'sub_default'}</span>
                      <div style={{ display: 'flex', gap: '16px', alignItems: 'center' }}>
                        {!adminSubmissions.has_private_key && (
                          <span style={{ color: 'var(--accent-red)', fontWeight: 500 }}>Key Required</span>
                        )}
                        {(adminSubmissions.submissions?.length || 0) > 0 && (
                          <button
                            onClick={() => triggerAdminConfirmClearSubmissions(activeAdminDrivePath)}
                            disabled={evaluatingTarget !== null}
                            style={{
                              background: 'none',
                              border: 'none',
                              color: evaluatingTarget !== null ? 'var(--text-muted)' : 'var(--accent-red)',
                              fontSize: '0.74rem',
                              fontFamily: 'var(--font-sans)',
                              fontWeight: 600,
                              cursor: evaluatingTarget !== null ? 'not-allowed' : 'pointer',
                              padding: 0
                            }}
                          >
                            Clear Submissions
                          </button>
                        )}
                      </div>
                    </div>
                  )}

                  {/* PROMPT FOR MISSING PRIVATE KEY */}
                  {showEvalKeyModal && (
                    <div style={{ marginBottom: '16px', display: 'flex', flexDirection: 'column', gap: '8px' }}>
                      <div style={{ fontSize: '0.78rem', color: 'var(--text-secondary)' }}>
                        Save Private Key for <code>{adminSubmissions.submission_id}</code>:
                      </div>
                      <input
                        type="password"
                        value={instructorPrivateKey}
                        onChange={(e) => setInstructorPrivateKey(e.target.value)}
                        placeholder="Paste Base64 X25519 Private Key..."
                        style={{
                          padding: '6px 10px',
                          fontSize: '0.8rem',
                          fontFamily: 'var(--font-mono)',
                          backgroundColor: 'transparent',
                          border: '1px solid var(--border-color)',
                          borderRadius: '6px',
                          color: 'var(--text-primary)',
                          outline: 'none'
                        }}
                      />
                      <div style={{ display: 'flex', gap: '8px', justifyContent: 'flex-end' }}>
                        <button
                          onClick={() => setShowEvalKeyModal(false)}
                          style={{
                            padding: '4px 10px',
                            fontSize: '0.75rem',
                            backgroundColor: 'transparent',
                            color: 'var(--text-secondary)',
                            border: 'none',
                            cursor: 'pointer'
                          }}
                        >
                          Cancel
                        </button>
                        <button
                          onClick={handleSaveMissingKey}
                          disabled={!instructorPrivateKey.trim() || evaluatingTarget !== null}
                          style={{
                            padding: '5px 12px',
                            fontSize: '0.78rem',
                            fontWeight: 600,
                            backgroundColor: 'var(--btn-primary-bg)',
                            color: 'var(--btn-primary-text)',
                            border: '1px solid var(--border-color)',
                            borderRadius: '6px',
                            cursor: instructorPrivateKey.trim() && evaluatingTarget === null ? 'pointer' : 'not-allowed'
                          }}
                        >
                          Save Key & Evaluate
                        </button>
                      </div>
                    </div>
                  )}

                  {!adminSubmissions.prepared ? (
                    <div className="empty-recents-msg" style={{ paddingLeft: 0, fontStyle: 'normal' }}>
                      <div style={{ marginBottom: '10px', color: 'var(--text-secondary)', fontSize: '0.85rem' }}>
                        Submissions directory is not prepared.
                      </div>

                      <button
                        onClick={() => handlePrepareAdminSubmission(activeAdminDrivePath)}
                        style={{
                          background: 'none',
                          border: 'none',
                          color: 'var(--primary)',
                          fontSize: '0.8rem',
                          fontWeight: 600,
                          cursor: 'pointer',
                          padding: 0
                        }}
                      >
                        + Prepare Submission Directory
                      </button>
                    </div>
                  ) : (
                    <div className="vscode-recents-list">
                      {(adminSubmissions.submissions || []).length === 0 ? (
                        <div className="empty-recents-msg">No student packages submitted yet.</div>
                      ) : (
                        (adminSubmissions.submissions || []).map((sub, idx) => {
                          const targetKey = sub.rel_path || sub.filename;
                          const isEvaluatingThis = evaluatingTarget === targetKey;
                          const isAnyEvaluating = evaluatingTarget !== null;

                          return (
                            <div key={idx} className="vscode-recent-item" style={{ justifyContent: 'space-between' }}>
                              <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                                <span className="recent-icon"><LockIcon size={14} /></span>
                                <div className="recent-details">
                                  <span className="recent-name" style={{ fontFamily: 'var(--font-mono)', fontSize: '0.82rem' }}>
                                    {sub.filename}
                                  </span>
                                  <span className="recent-path" style={{ fontSize: '0.7rem' }}>
                                    {formatBytes(sub.size)}
                                  </span>
                                </div>
                              </div>

                              <button
                                title="Evaluate individual submission"
                                onClick={() => startSingleEvaluation(targetKey)}
                                disabled={isAnyEvaluating}
                                style={{
                                  background: 'none',
                                  border: 'none',
                                  color: isEvaluatingThis ? 'var(--accent)' : (isAnyEvaluating ? 'var(--text-muted)' : 'var(--primary)'),
                                  fontSize: '0.75rem',
                                  fontWeight: 600,
                                  cursor: isAnyEvaluating ? 'not-allowed' : 'pointer'
                                }}
                              >
                                {isEvaluatingThis ? 'Evaluating...' : 'Evaluate'}
                              </button>
                            </div>
                          );
                        })
                      )}
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* EVALUATION RESULTS SUMMARY TABLE */}
            {evalResults && (
              <div className="vscode-section" style={{ marginTop: '24px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '14px', borderBottom: '1px solid var(--border-color)', paddingBottom: '6px' }}>
                  <h2 style={{ border: 'none', padding: 0, margin: 0 }}>
                    Evaluation Summary ({evalResults.length})
                  </h2>

                  <button
                    onClick={exportCSVReport}
                    style={{
                      background: 'none',
                      border: 'none',
                      color: 'var(--primary)',
                      fontSize: '0.78rem',
                      fontWeight: 600,
                      cursor: 'pointer',
                      padding: 0
                    }}
                  >
                    Export CSV Report
                  </button>
                </div>

                <div style={{ overflowX: 'auto' }}>
                  <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.8rem', textAlign: 'left' }}>
                    <thead>
                      <tr style={{ borderBottom: '1px solid var(--border-color)', color: 'var(--text-muted)' }}>
                        <th style={{ padding: '8px 12px' }}>Student ID</th>
                        <th style={{ padding: '8px 12px' }}>Lab ID</th>
                        <th style={{ padding: '8px 12px' }}>Version</th>
                        <th style={{ padding: '8px 12px' }}>Status</th>
                        <th style={{ padding: '8px 12px' }}>Score</th>
                      </tr>
                    </thead>
                    <tbody>
                      {evalResults.map((res, idx) => (
                        <tr key={idx} style={{ borderBottom: '1px solid var(--border-color)' }}>
                          <td style={{ padding: '8px 12px', fontFamily: 'var(--font-mono)', fontWeight: 600 }}>{res.student_id || 'Unknown'}</td>
                          <td style={{ padding: '8px 12px' }}>{res.lab_id || 'N/A'}</td>
                          <td style={{ padding: '8px 12px', fontFamily: 'var(--font-mono)' }}>{res.version || 'N/A'}</td>
                          <td style={{ padding: '8px 12px' }}>
                            <span style={{
                              fontSize: '0.72rem',
                              fontWeight: 700,
                              textTransform: 'uppercase',
                              color: res.status === 'passed' ? 'var(--accent)' : 'var(--accent-red)'
                            }}>
                              {res.status}
                            </span>
                          </td>
                          <td style={{ padding: '8px 12px', fontWeight: 600 }}>
                            {res.earned_points} / {res.max_points}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </div>
        )}
      </main>

      <CommandPalette />
    </div>
  );
}
