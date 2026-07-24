import React, { useState, useEffect, useRef } from 'react';
import { useStudent } from '../StudentContext';
import { DriveIcon, OpenFolderIcon, LockIcon, CloseIcon } from './Icons';

export default function SubmitWizard({ isOpen, onClose }) {
  const {
    activeWorkspacePath,
    remoteServers,
    remoteServerStatuses,
    recentDrives,
    addRemoteServer,
    checkRemoteServerHealth
  } = useStudent();

  // Wizard Steps: 'target' | 'add_server' | 'add_drive' | 'student' | 'new_student' | 'pin' | 'submitting' | 'result'
  const [step, setStep] = useState('target');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);

  // Selected Data
  const [selectedTarget, setSelectedTarget] = useState(null); // { type: 'remote'|'drive', target: string }
  const [selectedStudent, setSelectedStudent] = useState(null); // { student_id: string, org_id: string }

  // Form inputs
  const [serverInput, setServerInput] = useState('');
  const [driveInput, setDriveInput] = useState('');
  const [studentIdInput, setStudentIdInput] = useState('');
  const [orgIdInput, setOrgIdInput] = useState('default');
  const [pinInput, setPinInput] = useState('');
  const [newPinInput, setNewPinInput] = useState('');
  const [showNewPin, setShowNewPin] = useState(false);

  // Results & Errors
  const [validationError, setValidationError] = useState('');
  const [submitResult, setSubmitResult] = useState(null);
  const [driveSubmissionsStatus, setDriveSubmissionsStatus] = useState({}); // path -> prepared (bool)

  // Saved student profiles
  const [savedProfiles, setSavedProfiles] = useState(() => {
    try {
      const stored = localStorage.getItem('tdes_student_profiles');
      return stored ? JSON.parse(stored) : [];
    } catch {
      return [];
    }
  });

  const wizardRef = useRef(null);

  // Reset state when opening
  useEffect(() => {
    if (!isOpen) return;
    setStep('target');
    setSearchQuery('');
    setSelectedIndex(0);
    setSelectedTarget(null);
    setSelectedStudent(null);
    setPinInput('');
    setNewPinInput('');
    setShowNewPin(false);
    setValidationError('');
    setSubmitResult(null);

    // Fetch workspace submission config to preload
    fetch('/api/workspace/submit-config')
      .then(res => res.json())
      .then(data => {
        if (data.student_id) {
          // Prepopulate form fields
          setStudentIdInput(data.student_id);
          setOrgIdInput(data.org_id || 'default');
          
          // Seed saved profiles if empty
          setSavedProfiles(prev => {
            const hasProfile = prev.some(p => p.student_id === data.student_id && p.org_id === data.org_id);
            if (!hasProfile) {
              const updated = [{ student_id: data.student_id, org_id: data.org_id || 'default' }, ...prev];
              localStorage.setItem('tdes_student_profiles', JSON.stringify(updated));
              return updated;
            }
            return prev;
          });
        }
      })
      .catch(() => {});

    // Check drive validation for all recent drives
    recentDrives.forEach(drivePath => {
      fetch(`/api/drive/submissions?path=${encodeURIComponent(drivePath)}`)
        .then(res => res.json())
        .then(data => {
          setDriveSubmissionsStatus(prev => ({
            ...prev,
            [drivePath]: !!data.prepared
          }));
        })
        .catch(() => {});
    });
  }, [isOpen, recentDrives]);

  if (!isOpen) return null;

  // Filter Target Options (Hide offline remote servers and unprepared drives)
  const getFilteredTargets = () => {
    const list = [];

    // Option: Add new remote server
    list.push({
      type: 'action',
      action: 'add_server',
      label: '+ Add new remote server...',
      desc: 'Register and validate a new TDES evaluation server URL'
    });

    // Option: Add new drive path
    list.push({
      type: 'action',
      action: 'add_drive',
      label: '+ Add new drive path...',
      desc: 'Browse or register a local USB/disk directory'
    });

    // Filtered online remote servers
    remoteServers.forEach(serverUrl => {
      const status = remoteServerStatuses[serverUrl];
      if (status && status.online) {
        list.push({
          type: 'target_remote',
          target: serverUrl,
          label: serverUrl,
          desc: 'Remote evaluation server (ONLINE)'
        });
      }
    });

    // Filtered ready drives
    recentDrives.forEach(drivePath => {
      const isPrepared = driveSubmissionsStatus[drivePath];
      if (isPrepared) {
        list.push({
          type: 'target_drive',
          target: drivePath,
          label: drivePath.split('/').pop() || drivePath.split('\\').pop(),
          desc: drivePath
        });
      }
    });

    // Apply search query filtering
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      return list.filter(item => 
        item.label.toLowerCase().includes(q) || 
        (item.desc && item.desc.toLowerCase().includes(q))
      );
    }
    return list;
  };

  // Filter Student Profile Options
  const getFilteredStudents = () => {
    const list = [];
    list.push({
      type: 'action',
      action: 'new_profile',
      label: '+ Create new student profile...',
      desc: 'Specify a new Student ID and Organization identifier'
    });

    savedProfiles.forEach(p => {
      list.push({
        type: 'profile',
        profile: p,
        label: `${p.student_id} (Org: ${p.org_id})`,
        desc: 'Select saved profile credentials'
      });
    });

    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      return list.filter(item => 
        item.label.toLowerCase().includes(q) || 
        (item.desc && item.desc.toLowerCase().includes(q))
      );
    }
    return list;
  };

  const handleTargetSelect = (item) => {
    if (item.action === 'add_server') {
      setStep('add_server');
      setSearchQuery('');
      setServerInput('');
      setSelectedIndex(0);
      setValidationError('');
    } else if (item.action === 'add_drive') {
      setStep('add_drive');
      setSearchQuery('');
      setDriveInput('');
      setSelectedIndex(0);
      setValidationError('');
    } else if (item.type === 'target_remote') {
      setSelectedTarget({ type: 'remote', target: item.target });
      setStep('student');
      setSearchQuery('');
      setSelectedIndex(0);
      setValidationError('');
    } else if (item.type === 'target_drive') {
      setSelectedTarget({ type: 'drive', target: item.target });
      setStep('student');
      setSearchQuery('');
      setSelectedIndex(0);
      setValidationError('');
    }
  };

  const handleStudentSelect = (item) => {
    if (item.action === 'new_profile') {
      setStep('new_student');
      setStudentIdInput('');
      setOrgIdInput('default');
      setValidationError('');
    } else if (item.type === 'profile') {
      setSelectedStudent(item.profile);
      setStep('pin');
      setValidationError('');
    }
  };

  const handleAddServerConfirm = async () => {
    const cleanUrl = serverInput.trim();
    if (!cleanUrl) return;

    setValidationError('');
    try {
      const res = await fetch(`/api/remote/health?url=${encodeURIComponent(cleanUrl)}`);
      if (res.ok) {
        const data = await res.json();
        if (data.online) {
          addRemoteServer(cleanUrl);
          setSelectedTarget({ type: 'remote', target: cleanUrl });
          setStep('student');
          setSearchQuery('');
          setSelectedIndex(0);
        } else {
          setValidationError(`Server ${cleanUrl} is offline or unreachable: ${data.error || 'Connection failed'}`);
        }
      } else {
        setValidationError('Failed to validate remote registry server health');
      }
    } catch (err) {
      setValidationError('Connection error: ' + err.message);
    }
  };

  const handleAddDriveConfirm = async () => {
    const cleanPath = driveInput.trim();
    if (!cleanPath) return;

    setValidationError('');
    try {
      const res = await fetch(`/api/drive/submissions?path=${encodeURIComponent(cleanPath)}`);
      if (res.ok) {
        const data = await res.json();
        if (data.prepared) {
          // Add to recent drives via context side-effect or state
          setDriveSubmissionsStatus(prev => ({ ...prev, [cleanPath]: true }));
          setSelectedTarget({ type: 'drive', target: cleanPath });
          setStep('student');
          setSearchQuery('');
          setSelectedIndex(0);
        } else {
          setValidationError(`Submission directory not enabled/prepared on path: ${cleanPath}`);
        }
      } else {
        setValidationError('Failed to validate local drive path');
      }
    } catch (err) {
      setValidationError('Error: ' + err.message);
    }
  };

  const handleNewStudentConfirm = () => {
    const cleanId = studentIdInput.trim();
    const cleanOrg = orgIdInput.trim() || 'default';
    if (!cleanId) {
      setValidationError('Student ID is required');
      return;
    }

    const profile = { student_id: cleanId, org_id: cleanOrg };
    setSelectedStudent(profile);

    // Save profile to list
    setSavedProfiles(prev => {
      const hasProfile = prev.some(p => p.student_id === cleanId && p.org_id === cleanOrg);
      if (!hasProfile) {
        const updated = [profile, ...prev];
        localStorage.setItem('tdes_student_profiles', JSON.stringify(updated));
        return updated;
      }
      return prev;
    });

    setStep('pin');
    setValidationError('');
  };

  const handleFinalSubmit = async () => {
    if (selectedTarget.type === 'remote' && !pinInput.trim()) {
      setValidationError('Security PIN code is required');
      return;
    }

    setValidationError('');
    setStep('submitting');

    try {
      const res = await fetch('/api/workspace/submit', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: activeWorkspacePath,
          strategy: selectedTarget.type,
          target: selectedTarget.target,
          student_id: selectedStudent.student_id,
          org_id: selectedStudent.org_id,
          pin: pinInput.trim(),
          new_pin: showNewPin ? newPinInput.trim() : ''
        })
      });

      const data = await res.json();
      if (!res.ok) {
        setStep('pin');
        setValidationError(data.error || 'Submission failed');
      } else {
        setSubmitResult(data.result);
        setStep('result');
      }
    } catch (err) {
      setStep('pin');
      setValidationError('Connection error: ' + err.message);
    }
  };

  // Keyboard Navigation Handlers
  const handleKeyDown = (e) => {
    if (e.key === 'Escape') {
      onClose();
      return;
    }

    const items = step === 'target' ? getFilteredTargets() : step === 'student' ? getFilteredStudents() : [];

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelectedIndex(prev => (prev + 1) % Math.max(1, items.length));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelectedIndex(prev => (prev - 1 + items.length) % Math.max(1, items.length));
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (step === 'target' && items[selectedIndex]) {
        handleTargetSelect(items[selectedIndex]);
      } else if (step === 'student' && items[selectedIndex]) {
        handleStudentSelect(items[selectedIndex]);
      } else if (step === 'add_server') {
        handleAddServerConfirm();
      } else if (step === 'add_drive') {
        handleAddDriveConfirm();
      } else if (step === 'new_student') {
        handleNewStudentConfirm();
      } else if (step === 'pin') {
        handleFinalSubmit();
      }
    }
  };

  // Render evaluation response nicely
  const parseResultView = (resultStr) => {
    if (!resultStr) return null;
    try {
      const parsed = JSON.parse(resultStr);
      if (parsed && parsed.earned_points !== undefined) {
        return (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
            <div style={{ fontSize: '1.0rem', fontWeight: 700, color: 'var(--accent)', marginBottom: '8px' }}>
              Evaluation Completed: {parsed.earned_points} / {parsed.max_points} points
            </div>
            {parsed.results && parsed.results.map((tr, idx) => (
              <div key={idx} style={{ padding: '6px 10px', borderRadius: '4px', backgroundColor: 'var(--bg-main)', border: '1px solid var(--border-color)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.78rem', fontFamily: 'var(--font-mono)' }}>
                <span>{tr.command}</span>
                <span style={{ color: tr.status === 'pass' ? 'var(--accent)' : 'var(--accent-red)', fontWeight: 600 }}>
                  {tr.status === 'pass' ? `PASSED (${tr.points_earned}/${tr.points_possible})` : 'FAILED'}
                </span>
              </div>
            ))}
          </div>
        );
      }
    } catch {}
    return (
      <div style={{ fontSize: '0.8rem', fontFamily: 'var(--font-mono)', whiteSpace: 'pre-wrap', color: 'var(--text-primary)' }}>
        {resultStr}
      </div>
    );
  };

  return (
    <div className="quick-open-overlay-blur" style={{ zIndex: 120 }}>
      <div 
        className="quick-open-container" 
        ref={wizardRef} 
        onKeyDown={handleKeyDown}
        style={{
          width: '580px',
          maxWidth: '94vw',
          backgroundColor: 'var(--bg-card)',
          border: '1px solid var(--border-color)',
          borderRadius: '8px',
          boxShadow: '0 16px 48px rgba(0,0,0,0.3)',
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden'
        }}
      >
        {/* Wizard Header Bar */}
        <div style={{ padding: '10px 14px', fontSize: '0.72rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid var(--border-color)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', backgroundColor: 'var(--bg-card)' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <span>SUBMIT EXERCISE WIZARD</span>
          </div>
          <button 
            onClick={onClose} 
            style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-muted)', display: 'flex', alignItems: 'center' }}
          >
            <CloseIcon size={14} />
          </button>
        </div>

        {/* Selected Flow Header Indicator */}
        {(selectedTarget || selectedStudent) && (
          <div style={{ padding: '10px 14px', borderBottom: '1px solid var(--border-color)', backgroundColor: 'var(--bg-main)', fontSize: '0.75rem', color: 'var(--text-muted)', display: 'flex', flexWrap: 'wrap', gap: '16px' }}>
            {selectedTarget && (
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: '4px' }}>
                <strong>Destination:</strong> {selectedTarget.type === 'remote' ? 'Remote Server:' : 'Local Drive:'} {selectedTarget.target}
              </span>
            )}
            {selectedStudent && (
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: '4px' }}>
                <strong>Student Profile:</strong> {selectedStudent.student_id} (Org: {selectedStudent.org_id})
              </span>
            )}
          </div>
        )}

        {/* Input Row / Form Area */}
        <div className="quick-open-input-row" style={{ borderBottom: step === 'result' ? 'none' : '1px solid var(--border-color)' }}>
          {step === 'target' && (
            <input 
              type="text" 
              value={searchQuery} 
              onChange={(e) => { setSearchQuery(e.target.value); setSelectedIndex(0); }}
              placeholder="Search or add target connection..."
              className="quick-open-input"
              autoFocus
            />
          )}

          {step === 'student' && (
            <input 
              type="text" 
              value={searchQuery} 
              onChange={(e) => { setSearchQuery(e.target.value); setSelectedIndex(0); }}
              placeholder="Search or create student credentials profile..."
              className="quick-open-input"
              autoFocus
            />
          )}

          {step === 'add_server' && (
            <div style={{ display: 'flex', flex: 1, gap: '8px', padding: '0 8px' }}>
              <input 
                type="text" 
                value={serverInput} 
                onChange={(e) => setServerInput(e.target.value)}
                placeholder="Type remote registry server URL and press Enter..."
                className="quick-open-input"
                style={{ flex: 1 }}
                autoFocus
              />
              <button 
                onClick={handleAddServerConfirm}
                className="vscode-theme-text-btn"
                style={{ alignSelf: 'center', backgroundColor: 'var(--primary)', color: '#fff', border: 'none', padding: '4px 12px' }}
              >
                Validate
              </button>
            </div>
          )}

          {step === 'add_drive' && (
            <div style={{ display: 'flex', flex: 1, gap: '8px', padding: '0 8px' }}>
              <input 
                type="text" 
                value={driveInput} 
                onChange={(e) => setDriveInput(e.target.value)}
                placeholder="Type drive/USB path and press Enter..."
                className="quick-open-input"
                style={{ flex: 1 }}
                autoFocus
              />
              <button 
                onClick={handleAddDriveConfirm}
                className="vscode-theme-text-btn"
                style={{ alignSelf: 'center', backgroundColor: 'var(--primary)', color: '#fff', border: 'none', padding: '4px 12px' }}
              >
                Validate
              </button>
            </div>
          )}

          {step === 'new_student' && (
            <div style={{ display: 'flex', flex: 1, gap: '10px', padding: '0 10px', alignItems: 'center' }}>
              <input 
                type="text" 
                value={studentIdInput} 
                onChange={(e) => setStudentIdInput(e.target.value)}
                placeholder="Student ID *"
                className="quick-open-input"
                style={{ flex: 2, padding: '4px 6px', fontSize: '0.8rem' }}
                autoFocus
              />
              <input 
                type="text" 
                value={orgIdInput} 
                onChange={(e) => setOrgIdInput(e.target.value)}
                placeholder="Org ID (default: default)"
                className="quick-open-input"
                style={{ flex: 1, padding: '4px 6px', fontSize: '0.8rem' }}
              />
              <button 
                onClick={handleNewStudentConfirm}
                className="vscode-theme-text-btn"
                style={{ backgroundColor: 'var(--primary)', color: '#fff', border: 'none', padding: '4px 12px' }}
              >
                Next
              </button>
            </div>
          )}

          {step === 'pin' && (
            <div style={{ display: 'flex', flex: 1, gap: '12px', padding: '0 12px', alignItems: 'center' }}>
              {selectedTarget.type === 'remote' ? (
                <>
                  <input 
                    type="password" 
                    value={pinInput} 
                    onChange={(e) => setPinInput(e.target.value)}
                    placeholder="Enter Security PIN *"
                    className="quick-open-input"
                    style={{ flex: 2 }}
                    autoFocus
                  />
                  {showNewPin ? (
                    <input 
                      type="password" 
                      value={newPinInput} 
                      onChange={(e) => setNewPinInput(e.target.value)}
                      placeholder="Enter new PIN"
                      className="quick-open-input"
                      style={{ flex: 2 }}
                    />
                  ) : (
                    <button 
                      onClick={() => setShowNewPin(true)}
                      className="vscode-theme-text-btn"
                      style={{ fontSize: '0.7rem', whiteSpace: 'nowrap' }}
                    >
                      Reset PIN
                    </button>
                  )}
                </>
              ) : (
                <div style={{ flex: 1, fontSize: '0.82rem', color: 'var(--text-secondary)', fontWeight: 500 }}>
                  Ready to copy package to drive. No PIN required.
                </div>
              )}
              <button 
                onClick={handleFinalSubmit}
                className="vscode-theme-text-btn"
                style={{ backgroundColor: 'var(--btn-primary-bg)', color: 'var(--btn-primary-text)', border: 'none', padding: '6px 14px', fontWeight: 600 }}
              >
                Submit Now
              </button>
            </div>
          )}

          {step === 'submitting' && (
            <div style={{ display: 'flex', flex: 1, padding: '0 12px', alignItems: 'center', justifyContent: 'center', gap: '8px', color: 'var(--text-secondary)', fontSize: '0.82rem' }}>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" style={{ animation: 'spin 1.5s linear infinite' }}>
                <circle cx="12" cy="12" r="10" strokeDasharray="30" strokeDashoffset="10" />
              </svg>
              <span>Submitting package and executing remote grading targets...</span>
            </div>
          )}

          {step === 'result' && (
            <div style={{ flex: 1, fontSize: '0.82rem', padding: '0 10px', color: 'var(--accent)', fontWeight: 600, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span>Submission complete! See evaluation feedback below:</span>
              <button 
                onClick={onClose}
                className="vscode-theme-text-btn"
                style={{ backgroundColor: 'var(--primary)', color: '#fff', border: 'none', padding: '4px 12px' }}
              >
                Close Wizard
              </button>
            </div>
          )}
        </div>

        {/* Validation Errors */}
        {validationError && (
          <div className="validation-alert-error" style={{ margin: '10px 14px 4px 14px', textAlign: 'left' }}>
            {validationError}
          </div>
        )}

        {/* Results List / Details */}
        <div style={{ maxHeight: '320px', overflowY: 'auto', display: 'flex', flexDirection: 'column' }}>
          {step === 'target' && (
            getFilteredTargets().map((item, idx) => (
              <div 
                key={idx}
                className={`quick-open-item ${selectedIndex === idx ? 'active' : ''}`}
                onClick={() => handleTargetSelect(item)}
                onMouseEnter={() => setSelectedIndex(idx)}
                style={{ borderBottom: '1px solid var(--border-color)' }}
              >
                <div className="recent-details">
                  <span className="recent-name" style={{ fontSize: '0.82rem', display: 'flex', alignItems: 'center', gap: '6px' }}>
                    {item.label}
                  </span>
                  {item.desc && (
                    <span className="recent-path" style={{ fontSize: '0.7rem', opacity: 0.6 }}>
                      {item.desc}
                    </span>
                  )}
                </div>
              </div>
            ))
          )}

          {step === 'student' && (
            getFilteredStudents().map((item, idx) => (
              <div 
                key={idx}
                className={`quick-open-item ${selectedIndex === idx ? 'active' : ''}`}
                onClick={() => handleStudentSelect(item)}
                onMouseEnter={() => setSelectedIndex(idx)}
                style={{ borderBottom: '1px solid var(--border-color)' }}
              >
                <div className="recent-details">
                  <span className="recent-name" style={{ fontSize: '0.82rem' }}>
                    {item.label}
                  </span>
                  {item.desc && (
                    <span className="recent-path" style={{ fontSize: '0.7rem', opacity: 0.6 }}>
                      {item.desc}
                    </span>
                  )}
                </div>
              </div>
            ))
          )}

          {step === 'result' && (
            <div style={{ padding: '16px 20px', textAlign: 'left', backgroundColor: 'var(--bg-main)' }}>
              {parseResultView(submitResult)}
            </div>
          )}

          {(step === 'add_server' || step === 'add_drive' || step === 'new_student' || step === 'pin') && (
            <div style={{ padding: '24px', textAlign: 'center', color: 'var(--text-muted)', fontSize: '0.75rem', fontStyle: 'italic' }}>
              Press ESC to cancel or Enter/Validate to proceed.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
