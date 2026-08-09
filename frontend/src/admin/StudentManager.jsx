import React, { useState, useEffect } from 'react';
import { useStudent } from '../student/StudentContext';

export default function StudentManager({ onBack }) {
  const { activeRegistryServer } = useStudent();
  const serverURL = activeRegistryServer?.url || 'http://localhost:8080';
  const bearerToken = activeRegistryServer?.token || '';

  const [students, setStudents] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [searchQuery, setSearchQuery] = useState('');

  // Single student modal state
  const [showSingleModal, setShowSingleModal] = useState(false);
  const [singleStudentID, setSingleStudentID] = useState('');
  const [singleOrgID, setSingleOrgID] = useState('default');
  const [singleOnboarding, setSingleOnboarding] = useState(false);
  const [singleError, setSingleError] = useState('');

  // Bulk CSV uploader state
  const [selectedFile, setSelectedFile] = useState(null);
  const [bulkOnboarding, setBulkOnboarding] = useState(false);
  const [bulkResult, setBulkResult] = useState(null);
  const [bulkError, setBulkError] = useState('');

  const fetchStudents = async () => {
    if (!serverURL) return;
    setLoading(true);
    setError('');
    try {
      const queryParams = new URLSearchParams({
        remote_url: serverURL,
        bearer_token: bearerToken,
      });
      const resp = await fetch(`/api/server/students?${queryParams.toString()}`);
      if (!resp.ok) {
        const data = await resp.json();
        throw new Error(data.error || `HTTP ${resp.status}`);
      }
      const data = await resp.json();
      setStudents(Array.isArray(data) ? data : []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStudents();
  }, [serverURL, bearerToken]);

  const handleSingleOnboardSubmit = async (e) => {
    e.preventDefault();
    if (!singleStudentID.trim() || !serverURL) return;

    setSingleOnboarding(true);
    setSingleError('');

    const csvContent = `student_id,org_id\n${singleStudentID.trim()},${singleOrgID.trim() || 'default'}`;
    const blob = new Blob([csvContent], { type: 'text/csv' });
    const formData = new FormData();
    formData.append('remote_url', serverURL);
    formData.append('bearer_token', bearerToken);
    formData.append('roster_csv', blob, 'single_onboard.csv');

    try {
      const resp = await fetch('/api/server/onboard', {
        method: 'POST',
        body: formData,
      });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || `HTTP ${resp.status}`);
      }
      setSingleStudentID('');
      setShowSingleModal(false);
      await fetchStudents();
    } catch (err) {
      setSingleError(err.message);
    } finally {
      setSingleOnboarding(false);
    }
  };

  const handleBulkOnboardSubmit = async (e) => {
    e.preventDefault();
    if (!selectedFile || !serverURL) return;

    setBulkOnboarding(true);
    setBulkResult(null);
    setBulkError('');

    const formData = new FormData();
    formData.append('remote_url', serverURL);
    formData.append('bearer_token', bearerToken);
    formData.append('roster_csv', selectedFile);

    try {
      const resp = await fetch('/api/server/onboard', {
        method: 'POST',
        body: formData,
      });
      const data = await resp.json();
      if (!resp.ok) {
        throw new Error(data.error || `HTTP ${resp.status}`);
      }
      setBulkResult(data);
      setSelectedFile(null);
      await fetchStudents();
    } catch (err) {
      setBulkError(err.message);
    } finally {
      setBulkOnboarding(false);
    }
  };

  const filteredStudents = students.filter(s => {
    const q = searchQuery.toLowerCase().trim();
    if (!q) return true;
    return (s.student_id || '').toLowerCase().includes(q) || (s.org_id || '').toLowerCase().includes(q);
  });

  return (
    <div className="vscode-welcome-container" style={{ textAlign: 'left', maxWidth: '1100px', margin: '0 auto' }}>
      {/* HEADER */}
      <header className="vscode-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1>Student Management</h1>
          <p style={{ fontFamily: 'var(--font-mono)', fontSize: '0.82rem', marginTop: '4px', opacity: 0.8 }}>
            Rostered Student Credentials & TOFU Activation State ({students.length} Total)
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
          placeholder="Filter by Student ID or Org ID..."
          style={{
            padding: '6px 12px',
            fontSize: '0.8rem',
            backgroundColor: 'transparent',
            border: '1px solid var(--border-color)',
            borderRadius: '4px',
            color: 'var(--text-primary)',
            width: '280px',
            outline: 'none'
          }}
        />

        <div style={{ display: 'flex', gap: '16px', alignItems: 'center' }}>
          <button
            onClick={() => setShowSingleModal(true)}
            style={{ background: 'none', border: 'none', color: 'var(--primary)', fontSize: '0.8rem', cursor: 'pointer', fontWeight: 600 }}
          >
            + Onboard Student
          </button>
          <button
            onClick={fetchStudents}
            disabled={loading}
            style={{ background: 'none', border: 'none', color: 'var(--primary)', fontSize: '0.8rem', cursor: 'pointer', fontWeight: 600 }}
          >
            {loading ? 'Refreshing...' : 'Refresh'}
          </button>
        </div>
      </div>

      {error && (
        <div className="validation-alert-error" style={{ marginBottom: '16px' }}>
          Error loading student roster: {error}
        </div>
      )}

      {/* ROSTER TABLE */}
      <div className="vscode-section" style={{ marginBottom: '28px' }}>
        <h2>Rostered Student Credentials</h2>

        {filteredStudents.length === 0 ? (
          <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)', fontStyle: 'italic', padding: '16px 0' }}>
            {loading ? 'Fetching rostered students...' : (searchQuery ? 'No student matched your search query.' : 'No students onboarded yet.')}
          </div>
        ) : (
          <div style={{ overflowX: 'auto', marginTop: '12px' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.8rem', textAlign: 'left' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--border-color)', color: 'var(--text-muted)' }}>
                  <th style={{ padding: '8px 12px' }}>Student ID</th>
                  <th style={{ padding: '8px 12px' }}>Org ID</th>
                  <th style={{ padding: '8px 12px' }}>TOFU PIN Status</th>
                  <th style={{ padding: '8px 12px' }}>Onboarded At</th>
                  <th style={{ padding: '8px 12px' }}>Last Active</th>
                </tr>
              </thead>
              <tbody>
                {filteredStudents.map((stud, idx) => {
                  const isActivated = !!stud.pin_hash;
                  return (
                    <tr key={`${stud.org_id}-${stud.student_id}-${idx}`} style={{ borderBottom: '1px solid var(--border-color)' }}>
                      <td style={{ padding: '8px 12px', fontFamily: 'var(--font-mono)', fontWeight: 600 }}>{stud.student_id}</td>
                      <td style={{ padding: '8px 12px' }}>{stud.org_id}</td>
                      <td style={{ padding: '8px 12px' }}>
                        <span style={{
                          fontSize: '0.72rem',
                          fontWeight: 700,
                          padding: '2px 8px',
                          borderRadius: '4px',
                          textTransform: 'uppercase',
                          backgroundColor: isActivated ? 'rgba(16, 185, 129, 0.12)' : 'rgba(245, 158, 11, 0.12)',
                          color: isActivated ? 'var(--accent)' : 'var(--accent-yellow, #d97706)'
                        }}>
                          {isActivated ? '● Activated' : '● Pending First Use'}
                        </span>
                      </td>
                      <td style={{ padding: '8px 12px', fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                        {stud.created_at ? new Date(stud.created_at).toLocaleString() : 'N/A'}
                      </td>
                      <td style={{ padding: '8px 12px', fontSize: '0.75rem', color: 'var(--text-muted)' }}>
                        {stud.updated_at ? new Date(stud.updated_at).toLocaleString() : 'N/A'}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* BULK ROSTER CSV IMPORT SECTION */}
      <div className="vscode-section">
        <h2>Bulk Roster CSV Import</h2>
        <p style={{ fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '12px' }}>
          Upload a CSV file containing <code>student_id</code> and optional <code>org_id</code> headers. Existing student PINs will remain unchanged.
        </p>

        <form onSubmit={handleBulkOnboardSubmit} style={{ display: 'flex', gap: '12px', alignItems: 'center' }}>
          <input
            type="file"
            accept=".csv"
            onChange={(e) => setSelectedFile(e.target.files[0] || null)}
            style={{ fontSize: '0.8rem', color: 'var(--text-primary)' }}
          />

          <button
            type="submit"
            disabled={!selectedFile || bulkOnboarding}
            style={{
              padding: '6px 14px',
              fontSize: '0.78rem',
              fontWeight: 600,
              backgroundColor: 'var(--btn-primary-bg)',
              color: 'var(--btn-primary-text)',
              border: 'none',
              borderRadius: '4px',
              cursor: selectedFile && !bulkOnboarding ? 'pointer' : 'not-allowed'
            }}
          >
            {bulkOnboarding ? 'Onboarding...' : 'Onboard CSV Roster'}
          </button>
        </form>

        {bulkResult && (
          <div style={{ marginTop: '12px', fontSize: '0.8rem', color: 'var(--accent)' }}>
            ✓ Successfully onboarded {bulkResult.onboarded} student(s).
          </div>
        )}

        {bulkError && (
          <div className="validation-alert-error" style={{ marginTop: '12px' }}>
            Onboarding Error: {bulkError}
          </div>
        )}
      </div>

      {/* SINGLE STUDENT ONBOARD MODAL */}
      {showSingleModal && (
        <div style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          backgroundColor: 'rgba(0,0,0,0.6)',
          backdropFilter: 'blur(4px)',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          zIndex: 100
        }}>
          <div style={{
            backgroundColor: 'var(--bg-card)',
            border: '1px solid var(--border-color)',
            borderRadius: '8px',
            width: '420px',
            maxWidth: '90vw',
            padding: '20px',
            boxShadow: '0 12px 36px rgba(0,0,0,0.3)'
          }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '14px', borderBottom: '1px solid var(--border-color)', paddingBottom: '10px' }}>
              <h3 style={{ margin: 0, fontSize: '0.95rem', fontWeight: 700 }}>
                Onboard Single Student
              </h3>
              <button
                onClick={() => setShowSingleModal(false)}
                style={{ background: 'none', border: 'none', color: 'var(--text-muted)', fontSize: '1rem', cursor: 'pointer' }}
              >
                ✕
              </button>
            </div>

            <form onSubmit={handleSingleOnboardSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
              <div>
                <label style={{ fontSize: '0.74rem', fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>
                  Student ID <span style={{ color: 'var(--accent-red)' }}>*</span>
                </label>
                <input
                  type="text"
                  value={singleStudentID}
                  onChange={(e) => setSingleStudentID(e.target.value)}
                  placeholder="e.g. student-101"
                  required
                  style={{
                    width: '100%',
                    padding: '6px 10px',
                    fontSize: '0.8rem',
                    backgroundColor: 'transparent',
                    border: '1px solid var(--border-color)',
                    borderRadius: '4px',
                    color: 'var(--text-primary)',
                    outline: 'none'
                  }}
                />
              </div>

              <div>
                <label style={{ fontSize: '0.74rem', fontWeight: 600, color: 'var(--text-secondary)', display: 'block', marginBottom: '4px' }}>
                  Organization / Course ID
                </label>
                <input
                  type="text"
                  value={singleOrgID}
                  onChange={(e) => setSingleOrgID(e.target.value)}
                  placeholder="default"
                  style={{
                    width: '100%',
                    padding: '6px 10px',
                    fontSize: '0.8rem',
                    backgroundColor: 'transparent',
                    border: '1px solid var(--border-color)',
                    borderRadius: '4px',
                    color: 'var(--text-primary)',
                    outline: 'none'
                  }}
                />
              </div>

              {singleError && (
                <div style={{ fontSize: '0.75rem', color: 'var(--accent-red)' }}>
                  {singleError}
                </div>
              )}

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '12px', marginTop: '6px' }}>
                <button
                  type="button"
                  onClick={() => setShowSingleModal(false)}
                  style={{ background: 'none', border: 'none', color: 'var(--text-secondary)', fontSize: '0.8rem', cursor: 'pointer', fontWeight: 600 }}
                >
                  Cancel
                </button>

                <button
                  type="submit"
                  disabled={singleOnboarding || !singleStudentID.trim()}
                  style={{
                    padding: '6px 14px',
                    fontSize: '0.78rem',
                    fontWeight: 600,
                    backgroundColor: 'var(--btn-primary-bg)',
                    color: 'var(--btn-primary-text)',
                    border: 'none',
                    borderRadius: '4px',
                    cursor: 'pointer'
                  }}
                >
                  {singleOnboarding ? 'Onboarding...' : 'Onboard Student'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
