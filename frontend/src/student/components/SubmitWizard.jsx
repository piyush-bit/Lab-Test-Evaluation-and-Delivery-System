import React, { useState, useEffect, useRef } from 'react';
import { useStudent } from '../StudentContext';
import { CloseIcon } from './Icons';

export default function SubmitWizard({ isOpen, onClose }) {
  const {
    activeWorkspacePath,
    remoteServers,
    remoteServerStatuses,
    recentDrives,
    addRemoteServer
  } = useStudent();

  // Wizard Steps: 
  // 'target' | 'add_server' | 'add_drive' | 'student' | 'new_student_id' | 'new_student_org' | 'pin' | 'new_pin' | 'confirm_drive' | 'submitting' | 'result'
  const [step, setStep] = useState('target');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);

  // Selected Data
  const [selectedTarget, setSelectedTarget] = useState(null); // { type: 'remote'|'drive', target: string }
  const [selectedStudent, setSelectedStudent] = useState(null); // { student_id: string, org_id: string }

  // Folder Browsing State for 'add_drive'
  const [currentBrowsePath, setCurrentBrowsePath] = useState('');
  const [parentPath, setParentPath] = useState('');
  const [subDirectories, setSubDirectories] = useState([]);

  // Temp form fields
  const [tempStudentId, setTempStudentId] = useState('');
  const [pin, setPin] = useState('');
  const [newPin, setNewPin] = useState('');

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
    setPin('');
    setNewPin('');
    setValidationError('');
    setSubmitResult(null);

    // Fetch workspace submission config to preload
    fetch('/api/workspace/submit-config')
      .then(res => res.json())
      .then(data => {
        if (data.student_id) {
          // Prepopulate profiles
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

  // Load directories when entering folder browser mode
  useEffect(() => {
    if (step === 'add_drive') {
      fetchBrowseDirs(currentBrowsePath);
    }
  }, [step, currentBrowsePath]);

  const fetchBrowseDirs = async (path) => {
    try {
      const res = await fetch(`/api/browse?path=${encodeURIComponent(path)}`);
      if (res.ok) {
        const data = await res.json();
        setCurrentBrowsePath(data.current_path);
        setParentPath(data.parent_path || '');
        setSubDirectories(data.directories || []);
        setSelectedIndex(0);
        setValidationError('');
      } else {
        setValidationError('Failed to read directory');
      }
    } catch (err) {
      setValidationError('Error: ' + err.message);
    }
  };

  if (!isOpen) return null;

  // Filter Target Options (Hide offline remote servers and unprepared drives)
  const getFilteredTargets = () => {
    const list = [];

    list.push({
      type: 'action',
      action: 'add_server',
      label: '+ Add new remote server...',
      desc: 'Connect and register a new TDES evaluation server URL'
    });

    list.push({
      type: 'action',
      action: 'add_drive',
      label: '+ Add new drive path...',
      desc: 'Browse and open a local USB/disk directory'
    });

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
      desc: 'Specify a new Student ID and Organization ID'
    });

    savedProfiles.forEach(p => {
      list.push({
        type: 'profile',
        profile: p,
        label: `${p.student_id} (Org: ${p.org_id})`,
        desc: 'Use saved profile credentials'
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

  // Get current Folder Browser items
  const getBrowseItems = () => {
    const list = [];
    
    // Select current dir option
    list.push({
      type: 'select_dir',
      path: currentBrowsePath,
      label: `Select Current Folder: ${currentBrowsePath}`,
      desc: 'Verify and select this directory as submission target'
    });

    // Go up option
    if (parentPath) {
      list.push({
        type: 'up',
        path: parentPath,
        label: '.. (Go Up)',
        desc: `Navigate to parent: ${parentPath}`
      });
    }

    // Subdirectories
    subDirectories.forEach(dir => {
      list.push({
        type: 'dir',
        label: dir,
        desc: 'Navigate into folder'
      });
    });

    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      // Keep select_dir and up always, only filter directories
      return list.filter((item, index) => 
        index <= 1 || item.label.toLowerCase().includes(q)
      );
    }
    return list;
  };

  const handleTargetSelect = (item) => {
    if (item.action === 'add_server') {
      setStep('add_server');
      setSearchQuery('');
      setSelectedIndex(0);
      setValidationError('');
    } else if (item.action === 'add_drive') {
      setStep('add_drive');
      setSearchQuery('');
      setSelectedIndex(0);
      // Fetch default dir
      fetchBrowseDirs('');
    } else if (item.type === 'target_remote') {
      setSelectedTarget({ type: 'remote', target: item.target });
      setStep('student');
      setSearchQuery('');
      setSelectedIndex(0);
    } else if (item.type === 'target_drive') {
      setSelectedTarget({ type: 'drive', target: item.target });
      setStep('student');
      setSearchQuery('');
      setSelectedIndex(0);
    }
  };

  const handleBrowseSelect = async (item) => {
    if (item.type === 'select_dir') {
      setValidationError('');
      try {
        const res = await fetch(`/api/drive/submissions?path=${encodeURIComponent(item.path)}`);
        if (res.ok) {
          const data = await res.json();
          if (data.prepared) {
            setDriveSubmissionsStatus(prev => ({ ...prev, [item.path]: true }));
            setSelectedTarget({ type: 'drive', target: item.path });
            setStep('student');
            setSearchQuery('');
            setSelectedIndex(0);
          } else {
            setValidationError(`Submission directory not prepared on path: ${item.path}`);
          }
        } else {
          setValidationError('Failed to validate local drive path');
        }
      } catch (err) {
        setValidationError('Error: ' + err.message);
      }
    } else if (item.type === 'up') {
      setCurrentBrowsePath(item.path);
      setSearchQuery('');
    } else if (item.type === 'dir') {
      // separator-safe path join
      const separator = currentBrowsePath.includes('/') ? '/' : '\\';
      const cleanBase = currentBrowsePath.replace(/[/\\]$/, '');
      const newPath = `${cleanBase}${separator}${item.label}`;
      setCurrentBrowsePath(newPath);
      setSearchQuery('');
    }
  };

  const handleStudentSelect = (item) => {
    if (item.action === 'new_profile') {
      setStep('new_student_id');
      setSearchQuery('');
      setSelectedIndex(0);
    } else if (item.type === 'profile') {
      setSelectedStudent(item.profile);
      setStep(selectedTarget.type === 'remote' ? 'pin' : 'confirm_drive');
      setSearchQuery('');
      setSelectedIndex(0);
    }
  };

  const handleAddServerConfirm = async () => {
    const cleanUrl = searchQuery.trim();
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
          setValidationError(`Server is offline or unreachable: ${data.error || 'Connection failed'}`);
        }
      } else {
        setValidationError('Failed to validate remote registry server health');
      }
    } catch (err) {
      setValidationError('Connection error: ' + err.message);
    }
  };

  const handleFinalSubmit = async () => {
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
          pin: pin.trim(),
          new_pin: newPin.trim()
        })
      });

      const data = await res.json();
      if (!res.ok) {
        setStep(selectedTarget.type === 'remote' ? 'pin' : 'confirm_drive');
        setValidationError(data.error || 'Submission failed');
      } else {
        setSubmitResult(data.result);
        setStep('result');
      }
    } catch (err) {
      setStep(selectedTarget.type === 'remote' ? 'pin' : 'confirm_drive');
      setValidationError('Connection error: ' + err.message);
    }
  };

  const getConfirmItems = () => {
    const list = [
      {
        type: 'submit_action',
        action: 'submit',
        label: 'Submit'
      },
      {
        type: 'submit_action',
        action: 'cancel',
        label: 'Cancel'
      }
    ];

    if (step === 'pin') {
      list.splice(1, 0, {
        type: 'submit_action',
        action: 'change_pin',
        label: 'Change/Update Security PIN...'
      });
    }

    if (step === 'confirm_drive' && searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      return list.filter(item => 
        item.label.toLowerCase().includes(q) || 
        (item.desc && item.desc.toLowerCase().includes(q))
      );
    }

    return list;
  };

  const getResultOptions = () => {
    const list = [
      {
        action: 'close',
        label: 'Close',
        desc: 'Exit the submission wizard'
      }
    ];

    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      return list.filter(item => 
        item.label.toLowerCase().includes(q) || 
        (item.desc && item.desc.toLowerCase().includes(q))
      );
    }

    return list;
  };

  const handleCopyPath = () => {
    if (!submitResult) return;
    navigator.clipboard.writeText(submitResult);
    setValidationError('Copied to clipboard!');
    setTimeout(() => setValidationError(''), 2000);
  };

  // Keyboard navigation inside quick open
  const handleKeyDown = (e) => {
    if (e.key === 'Escape') {
      onClose();
      return;
    }

    const items = 
      step === 'target' ? getFilteredTargets() : 
      step === 'add_drive' ? getBrowseItems() : 
      step === 'student' ? getFilteredStudents() :
      (step === 'pin' || step === 'new_pin' || step === 'confirm_drive') ? getConfirmItems() :
      step === 'result' ? getResultOptions() : [];

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
      } else if (step === 'add_drive' && items[selectedIndex]) {
        handleBrowseSelect(items[selectedIndex]);
      } else if (step === 'add_server') {
        handleAddServerConfirm();
      } else if (step === 'student' && items[selectedIndex]) {
        handleStudentSelect(items[selectedIndex]);
      } else if (step === 'new_student_id') {
        const cleanId = searchQuery.trim();
        if (!cleanId) {
          setValidationError('Student ID is required');
          return;
        }
        setTempStudentId(cleanId);
        setStep('new_student_org');
        setSearchQuery('');
        setValidationError('');
      } else if (step === 'new_student_org') {
        const cleanOrg = searchQuery.trim() || 'default';
        const profile = { student_id: tempStudentId, org_id: cleanOrg };
        setSelectedStudent(profile);

        // Save profile locally
        setSavedProfiles(prev => {
          const hasProfile = prev.some(p => p.student_id === tempStudentId && p.org_id === cleanOrg);
          if (!hasProfile) {
            const updated = [profile, ...prev];
            localStorage.setItem('tdes_student_profiles', JSON.stringify(updated));
            return updated;
          }
          return prev;
        });

        setStep(selectedTarget.type === 'remote' ? 'pin' : 'confirm_drive');
        setSearchQuery('');
        setValidationError('');
      } else if ((step === 'pin' || step === 'new_pin' || step === 'confirm_drive') && items[selectedIndex]) {
        const actionItem = items[selectedIndex];
        if (actionItem.action === 'submit') {
          if (step === 'pin') {
            const cleanPin = searchQuery.trim();
            if (!cleanPin) {
              setValidationError('PIN code is required');
              return;
            }
            setPin(cleanPin);
          } else if (step === 'new_pin') {
            const cleanNewPin = searchQuery.trim();
            setNewPin(cleanNewPin);
          }
          handleFinalSubmit();
        } else if (actionItem.action === 'change_pin') {
          setStep('new_pin');
          setSearchQuery('');
          setSelectedIndex(0);
          setValidationError('');
        } else if (actionItem.action === 'cancel') {
          onClose();
        }
      } else if (step === 'result' && items[selectedIndex]) {
        const actionItem = items[selectedIndex];
        if (actionItem.action === 'copy') {
          handleCopyPath();
        } else if (actionItem.action === 'close') {
          onClose();
        }
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
            <div style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--accent)', marginBottom: '4px' }}>
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

    const isPath = resultStr.includes('/') || resultStr.includes('\\');
    if (isPath) {
      return (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
          <div>
            <div style={{ fontSize: '0.72rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '4px' }}>
              Status
            </div>
            <div style={{ fontSize: '0.82rem', color: 'var(--accent)', fontWeight: 600 }}>
              Submission successfully saved to disk
            </div>
          </div>
          <div>
            <div style={{ fontSize: '0.72rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '6px' }}>
              Manifest Path
            </div>
            <div style={{ 
              fontFamily: 'var(--font-mono)', 
              fontSize: '0.75rem', 
              color: 'var(--text-primary)', 
              backgroundColor: 'var(--bg-main)', 
              border: '1px solid var(--border-color)', 
              padding: '8px 12px', 
              borderRadius: '6px',
              wordBreak: 'break-all',
              lineHeight: '1.4'
            }}>
              {resultStr}
            </div>
          </div>
        </div>
      );
    }

    return (
      <div style={{ fontSize: '0.8rem', fontFamily: 'var(--font-mono)', whiteSpace: 'pre-wrap', color: 'var(--text-primary)' }}>
        {resultStr}
      </div>
    );
  };

  const formatSubmitResult = (resultStr) => {
    if (!resultStr) return null;
    const isPath = resultStr.includes('/') || resultStr.includes('\\');
    if (!isPath) return <span>Submitted successfully: {resultStr}</span>;

    const separator = resultStr.includes('/') ? '/' : '\\';
    const parts = resultStr.split(separator);
    const fileName = parts.pop();
    const dirPath = parts.join(separator) + separator;

    return (
      <span>
        Submitted as {dirPath}<strong>{fileName}</strong>
      </span>
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
        {((selectedTarget || selectedStudent) || step === 'result') && (
          <div style={{ padding: '10px 14px', borderBottom: '1px solid var(--border-color)', backgroundColor: 'var(--bg-main)', fontSize: '0.75rem', color: 'var(--text-muted)', wordBreak: 'break-all' }}>
            {step === 'result' ? (
              formatSubmitResult(submitResult)
            ) : (
              <span>Submitting as {selectedStudent ? `${selectedStudent.student_id}(org:${selectedStudent.org_id})` : '<id>(org:<org>)'} at {selectedTarget ? selectedTarget.target : '<location>'}</span>
            )}
          </div>
        )}

        {/* Input Row / Form Area */}
        {step !== 'submitting' && (
          <div className="quick-open-input-row" style={{ borderBottom: '1px solid var(--border-color)' }}>
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

            {step === 'add_server' && (
              <input 
                type="text" 
                value={searchQuery} 
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Type remote registry server URL and press Enter to connect..."
                className="quick-open-input"
                autoFocus
              />
            )}

            {step === 'add_drive' && (
              <input 
                type="text" 
                value={searchQuery} 
                onChange={(e) => {
                  setSearchQuery(e.target.value);
                  // Allow direct typing browsing path
                  if (e.target.value.includes('/') || e.target.value.includes('\\')) {
                    setCurrentBrowsePath(e.target.value);
                  }
                }}
                placeholder="Navigate or type drive folder path..."
                className="quick-open-input"
                autoFocus
              />
            )}

            {step === 'student' && (
              <input 
                type="text" 
                value={searchQuery} 
                onChange={(e) => { setSearchQuery(e.target.value); setSelectedIndex(0); }}
                placeholder="Search or select student profile..."
                className="quick-open-input"
                autoFocus
              />
            )}

            {step === 'new_student_id' && (
              <input 
                type="text" 
                value={searchQuery} 
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Type Student ID and press Enter..."
                className="quick-open-input"
                autoFocus
              />
            )}

            {step === 'new_student_org' && (
              <input 
                type="text" 
                value={searchQuery} 
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Type Organization ID (default: default) and press Enter..."
                className="quick-open-input"
                autoFocus
              />
            )}

            {step === 'pin' && (
              <input 
                type="password" 
                value={searchQuery} 
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Type Security PIN and press Enter..."
                className="quick-open-input"
                autoFocus
              />
            )}

            {step === 'new_pin' && (
              <input 
                type="password" 
                value={searchQuery} 
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Type new Security PIN and press Enter..."
                className="quick-open-input"
                autoFocus
              />
            )}

            {step === 'confirm_drive' && (
              <input 
                type="text" 
                value={searchQuery} 
                onChange={(e) => { setSearchQuery(e.target.value); setSelectedIndex(0); }}
                placeholder="Type 'Submit' or select option..."
                className="quick-open-input"
                autoFocus
              />
            )}

            {step === 'result' && (
              <input 
                type="text" 
                value={searchQuery} 
                onChange={(e) => { setSearchQuery(e.target.value); setSelectedIndex(0); }}
                placeholder="Type search or choose action..."
                className="quick-open-input"
                autoFocus
              />
            )}
          </div>
        )}

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

          {step === 'add_server' && (
            <div 
              className="quick-open-item active"
              onClick={handleAddServerConfirm}
              style={{ borderBottom: '1px solid var(--border-color)' }}
            >
              <div className="recent-details">
                <span className="recent-name" style={{ fontSize: '0.82rem' }}>
                  {searchQuery.trim() ? `Connect Server: ${searchQuery}` : 'Type Server URL and press Enter...'}
                </span>
                <span className="recent-path" style={{ fontSize: '0.7rem', opacity: 0.6 }}>
                  Registry URL will be validated for active health
                </span>
              </div>
            </div>
          )}

          {step === 'add_drive' && (
            getBrowseItems().map((item, idx) => (
              <div 
                key={idx}
                className={`quick-open-item ${selectedIndex === idx ? 'active' : ''}`}
                onClick={() => handleBrowseSelect(item)}
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

          {step === 'new_student_id' && (
            <div className="quick-open-item active">
              <div className="recent-details">
                <span className="recent-name" style={{ fontSize: '0.82rem' }}>
                  {searchQuery.trim() ? `Student ID: ${searchQuery}` : 'Type Student ID and press Enter...'}
                </span>
                <span className="recent-path" style={{ fontSize: '0.7rem', opacity: 0.6 }}>
                  Step 1: Set identifier
                </span>
              </div>
            </div>
          )}

          {step === 'new_student_org' && (
            <div className="quick-open-item active">
              <div className="recent-details">
                <span className="recent-name" style={{ fontSize: '0.82rem' }}>
                  {`Confirm Profile: ${tempStudentId} (Org: ${searchQuery.trim() || 'default'})`}
                </span>
                <span className="recent-path" style={{ fontSize: '0.7rem', opacity: 0.6 }}>
                  Step 2: Press Enter to save profile and proceed
                </span>
              </div>
            </div>
          )}

          {(step === 'pin' || step === 'new_pin' || step === 'confirm_drive') && (
            getConfirmItems().map((item, idx) => (
              <div 
                key={idx}
                className={`quick-open-item ${selectedIndex === idx ? 'active' : ''}`}
                onClick={() => {
                  if (item.action === 'submit') {
                    if (step === 'pin') {
                      const cleanPin = searchQuery.trim();
                      if (!cleanPin) {
                        setValidationError('PIN code is required');
                        return;
                      }
                      setPin(cleanPin);
                    } else if (step === 'new_pin') {
                      const cleanNewPin = searchQuery.trim();
                      setNewPin(cleanNewPin);
                    }
                    handleFinalSubmit();
                  } else if (item.action === 'change_pin') {
                    setStep('new_pin');
                    setSearchQuery('');
                    setSelectedIndex(0);
                    setValidationError('');
                  } else if (item.action === 'cancel') {
                    onClose();
                  }
                }}
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

          {step === 'submitting' && (
            <div style={{ padding: '24px', textAlign: 'center', color: 'var(--text-secondary)', fontSize: '0.8rem', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '10px' }}>
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" style={{ animation: 'spin 1.5s linear infinite', color: 'var(--primary)' }}>
                <circle cx="12" cy="12" r="10" strokeDasharray="30" strokeDashoffset="10" />
              </svg>
              <span>Submitting package and executing remote grading targets...</span>
            </div>
          )}

          {step === 'result' && (
            <>
              {getResultOptions().map((item, idx) => (
                <div 
                  key={idx}
                  className={`quick-open-item ${selectedIndex === idx ? 'active' : ''}`}
                  onClick={() => {
                    if (item.action === 'copy') {
                      handleCopyPath();
                    } else if (item.action === 'close') {
                      onClose();
                    }
                  }}
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
              ))}
              
              {submitResult && (() => {
                try {
                  const parsed = JSON.parse(submitResult);
                  if (parsed && parsed.earned_points !== undefined) {
                    return (
                      <div style={{ padding: '12px 14px', backgroundColor: 'var(--bg-main)', borderTop: '1px solid var(--border-color)', display: 'flex', flexDirection: 'column', gap: '6px' }}>
                        <div style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: '2px' }}>
                          Evaluation Feedback
                        </div>
                        {parsed.results && parsed.results.map((tr, idx) => (
                          <div key={idx} style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.75rem', fontFamily: 'var(--font-mono)' }}>
                            <span style={{ color: 'var(--text-secondary)' }}>{tr.command}</span>
                            <span style={{ color: tr.status === 'pass' ? 'var(--accent)' : 'var(--accent-red)', fontWeight: 600 }}>
                              {tr.status === 'pass' ? `PASSED (${tr.points_earned}/${tr.points_possible})` : 'FAILED'}
                            </span>
                          </div>
                        ))}
                      </div>
                    );
                  }
                } catch {}
                return null;
              })()}
            </>
          )}
        </div>
      </div>
    </div>
  );
}
