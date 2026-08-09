import React, { useState, useEffect } from 'react';
import { useStudent } from '../student/StudentContext';

export default function EvaluationsManager({ onBack }) {
  const { activeRegistryServer } = useStudent();
  const serverURL = activeRegistryServer?.url || 'http://localhost:8080';
  const bearerToken = activeRegistryServer?.token || '';

  const [evaluations, setEvaluations] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedEval, setSelectedEval] = useState(null);

  const fetchEvaluations = async () => {
    if (!serverURL) return;
    setLoading(true);
    setError('');
    try {
      const queryParams = new URLSearchParams({
        remote_url: serverURL,
        bearer_token: bearerToken,
      });
      const resp = await fetch(`/api/server/submissions?${queryParams.toString()}`);
      if (!resp.ok) {
        const data = await resp.json();
        throw new Error(data.error || `HTTP ${resp.status}`);
      }
      const data = await resp.json();
      setEvaluations(Array.isArray(data) ? data : []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchEvaluations();
  }, [serverURL, bearerToken]);

  const handleExportCSV = () => {
    if (!serverURL) return;
    const queryParams = new URLSearchParams({
      remote_url: serverURL,
      bearer_token: bearerToken,
      format: 'csv'
    });
    window.open(`/api/server/submissions?${queryParams.toString()}`, '_blank');
  };

  const filteredEvaluations = evaluations.filter(item => {
    const q = searchQuery.toLowerCase().trim();
    if (!q) return true;
    return (
      (item.student_id || '').toLowerCase().includes(q) ||
      (item.lab_id || '').toLowerCase().includes(q) ||
      (item.org_id || '').toLowerCase().includes(q) ||
      (item.status || '').toLowerCase().includes(q) ||
      (item.id || '').toLowerCase().includes(q)
    );
  });

  return (
    <div className="vscode-welcome-container" style={{ textAlign: 'left', maxWidth: '1100px', margin: '0 auto' }}>
      {/* HEADER */}
      <header className="vscode-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1>Recent Evaluations</h1>
          <p style={{ fontFamily: 'var(--font-mono)', fontSize: '0.82rem', marginTop: '4px', opacity: 0.8 }}>
            History of remote student submissions and grading results ({evaluations.length} Total)
          </p>
        </div>
        <button
          onClick={onBack}
          style={{ background: 'none', border: 'none', color: 'var(--text-secondary)', fontSize: '0.8rem', cursor: 'pointer', fontWeight: 600 }}
        >
          ← Back to Registry
        </button>
      </header>

      {/* TOOLBAR & SEARCH */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '24px', marginBottom: '16px' }}>
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder="Filter by Student ID, Lab ID, Org ID, or Status..."
          style={{
            padding: '6px 12px',
            fontSize: '0.8rem',
            backgroundColor: 'transparent',
            border: '1px solid var(--border-color)',
            borderRadius: '4px',
            color: 'var(--text-primary)',
            width: '340px',
            outline: 'none'
          }}
        />

        <div style={{ display: 'flex', gap: '16px', alignItems: 'center' }}>
          <button
            onClick={handleExportCSV}
            disabled={evaluations.length === 0}
            style={{
              background: 'none',
              border: 'none',
              color: evaluations.length > 0 ? 'var(--primary)' : 'var(--text-muted)',
              fontSize: '0.8rem',
              cursor: evaluations.length > 0 ? 'pointer' : 'not-allowed',
              fontWeight: 600
            }}
          >
            Export CSV
          </button>
          <button
            onClick={fetchEvaluations}
            disabled={loading}
            style={{ background: 'none', border: 'none', color: 'var(--primary)', fontSize: '0.8rem', cursor: 'pointer', fontWeight: 600 }}
          >
            {loading ? 'Refreshing...' : 'Refresh'}
          </button>
        </div>
      </div>

      {error && (
        <div className="validation-alert-error" style={{ marginBottom: '16px' }}>
          Error loading evaluation records: {error}
        </div>
      )}

      {/* EVALUATIONS TABLE */}
      <div className="vscode-section" style={{ marginBottom: '28px' }}>
        <h2>Evaluations Log</h2>

        {filteredEvaluations.length === 0 ? (
          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)', fontStyle: 'italic', padding: '16px 0' }}>
            {loading ? 'Fetching evaluations...' : (searchQuery ? 'No evaluation matched your filter query.' : 'No evaluation records found on registry server.')}
          </div>
        ) : (
          <div style={{ overflowX: 'auto', marginTop: '12px' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.8rem', textAlign: 'left' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--border-color)', color: 'var(--text-muted)' }}>
                  <th style={{ padding: '8px 12px' }}>Student ID</th>
                  <th style={{ padding: '8px 12px' }}>Lab ID & Version</th>
                  <th style={{ padding: '8px 12px' }}>Org ID</th>
                  <th style={{ padding: '8px 12px' }}>Status</th>
                  <th style={{ padding: '8px 12px' }}>Score</th>
                  <th style={{ padding: '8px 12px' }}>Submitted At</th>
                  <th style={{ padding: '8px 12px', textAlign: 'right' }}>Action</th>
                </tr>
              </thead>
              <tbody>
                {filteredEvaluations.map((item, idx) => {
                  const isCompleted = item.status === 'completed' || item.status === 'passed';
                  const isFailed = item.status === 'failed' || item.status === 'error';
                  const scoreStr = item.max_points !== undefined ? `${item.earned_points} / ${item.max_points}` : `${item.earned_points || 0} pts`;
                  
                  return (
                    <tr key={`${item.id || idx}`} style={{ borderBottom: '1px solid var(--border-color)' }}>
                      <td style={{ padding: '8px 12px', fontFamily: 'var(--font-mono)', fontWeight: 600 }}>
                        {item.student_id || 'N/A'}
                      </td>
                      <td style={{ padding: '8px 12px' }}>
                        <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 600 }}>{item.lab_id}</span>
                        {item.version && (
                          <span style={{ fontSize: '0.72rem', color: 'var(--text-muted)', marginLeft: '6px' }}>
                            v{item.version}
                          </span>
                        )}
                      </td>
                      <td style={{ padding: '8px 12px' }}>{item.org_id || 'default'}</td>
                      <td style={{ padding: '8px 12px' }}>
                        <span style={{
                          fontSize: '0.72rem',
                          fontWeight: 700,
                          padding: '2px 8px',
                          borderRadius: '4px',
                          textTransform: 'uppercase',
                          backgroundColor: isCompleted ? 'rgba(16, 185, 129, 0.12)' : (isFailed ? 'rgba(239, 68, 68, 0.12)' : 'rgba(245, 158, 11, 0.12)'),
                          color: isCompleted ? 'var(--accent)' : (isFailed ? 'var(--accent-red)' : 'var(--accent-yellow, #d97706)')
                        }}>
                          ● {item.status || 'processed'}
                        </span>
                      </td>
                      <td style={{ padding: '8px 12px', fontFamily: 'var(--font-mono)', fontWeight: 700 }}>
                        {scoreStr}
                      </td>
                      <td style={{ padding: '8px 12px', fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                        {item.created_at ? new Date(item.created_at).toLocaleString() : 'N/A'}
                      </td>
                      <td style={{ padding: '8px 12px', textAlign: 'right' }}>
                        <button
                          onClick={() => setSelectedEval(item)}
                          style={{
                            background: 'none',
                            border: '1px solid var(--border-color)',
                            borderRadius: '4px',
                            padding: '3px 8px',
                            color: 'var(--primary)',
                            fontSize: '0.75rem',
                            cursor: 'pointer',
                            fontWeight: 600
                          }}
                        >
                          Details
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* DETAILS MODAL */}
      {selectedEval && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundColor: 'rgba(0, 0, 0, 0.65)',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          zIndex: 1000
        }}>
          <div style={{
            backgroundColor: 'var(--bg-secondary, #1e1e1e)',
            border: '1px solid var(--border-color)',
            borderRadius: '6px',
            width: '600px',
            maxWidth: '90vw',
            maxHeight: '85vh',
            overflowY: 'auto',
            padding: '20px',
            boxShadow: '0 8px 32px rgba(0,0,0,0.5)'
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px', borderBottom: '1px solid var(--border-color)', paddingBottom: '12px' }}>
              <h3 style={{ margin: 0, fontSize: '1.05rem' }}>Evaluation Details</h3>
              <button
                onClick={() => setSelectedEval(null)}
                style={{ background: 'none', border: 'none', color: 'var(--text-muted)', fontSize: '1.2rem', cursor: 'pointer' }}
              >
                ✕
              </button>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '10px', fontSize: '0.82rem', marginBottom: '16px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: 'var(--text-muted)' }}>Student ID:</span>
                <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 600 }}>{selectedEval.student_id}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: 'var(--text-muted)' }}>Lab ID:</span>
                <span style={{ fontFamily: 'var(--font-mono)' }}>{selectedEval.lab_id} (v{selectedEval.version || '1.0.0'})</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: 'var(--text-muted)' }}>Organization:</span>
                <span>{selectedEval.org_id || 'default'}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: 'var(--text-muted)' }}>Total Score:</span>
                <span style={{ fontWeight: 700, color: 'var(--accent)' }}>
                  {selectedEval.earned_points} / {selectedEval.max_points} pts
                </span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: 'var(--text-muted)' }}>Timestamp:</span>
                <span>{selectedEval.created_at ? new Date(selectedEval.created_at).toLocaleString() : 'N/A'}</span>
              </div>
            </div>

            {selectedEval.results_json && (
              <div style={{ marginTop: '16px' }}>
                <h4 style={{ fontSize: '0.85rem', marginBottom: '8px', color: 'var(--text-secondary)' }}>Test Results Breakdown</h4>
                <pre style={{
                  backgroundColor: 'rgba(0,0,0,0.3)',
                  border: '1px solid var(--border-color)',
                  padding: '12px',
                  borderRadius: '4px',
                  fontSize: '0.75rem',
                  fontFamily: 'var(--font-mono)',
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-all',
                  maxHeight: '240px',
                  overflowY: 'auto'
                }}>
                  {(() => {
                    try {
                      return JSON.stringify(JSON.parse(selectedEval.results_json), null, 2);
                    } catch {
                      return selectedEval.results_json;
                    }
                  })()}
                </pre>
              </div>
            )}

            <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: '20px' }}>
              <button
                onClick={() => setSelectedEval(null)}
                style={{
                  padding: '6px 16px',
                  backgroundColor: 'var(--primary)',
                  color: '#fff',
                  border: 'none',
                  borderRadius: '4px',
                  fontSize: '0.8rem',
                  fontWeight: 600,
                  cursor: 'pointer'
                }}
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
