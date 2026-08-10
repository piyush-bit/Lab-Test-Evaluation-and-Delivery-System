import React, { useState, useEffect, useRef } from 'react';
import { useStudent } from '../StudentContext';
import { CloseIcon, BackIcon } from './Icons';

export default function SubmitWizard({ isOpen, onClose }) {
  const {
    activeWorkspacePath,
    remoteServers,
    remoteServerStatuses,
    remoteTokens,
    recentDrives,
    addRemoteServer
  } = useStudent();

  // Wizard Steps: 
  // 'target' | 'add_server' | 'add_server_token' | 'add_drive' | 'student' | 'new_student_id' | 'new_student_org' | 'pin' | 'new_pin' | 'confirm_drive' | 'processing' | 'error' | 'success'
  const [step, setStep] = useState('target');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);

  // Selected Data
  const [selectedTarget, setSelectedTarget] = useState(null); // { type: 'remote'|'drive', target: string, token?: string }
  const [pendingServerUrl, setPendingServerUrl] = useState('');
  const [selectedStudent, setSelectedStudent] = useState(null); // { student_id: string, org_id: string }

  // Folder Browsing State for 'add_drive'
  const [currentBrowsePath, setCurrentBrowsePath] = useState('');
  const [parentPath, setParentPath] = useState('');
  const [subDirectories, setSubDirectories] = useState([]);

  // Temp form fields
  const [tempStudentId, setTempStudentId] = useState('');
  const [pin, setPin] = useState('');
  const [newPin, setNewPin] = useState('');

  // Results, Errors & Expandable Logs
  const [validationError, setValidationError] = useState('');
  const [submitResult, setSubmitResult] = useState(null);
  const [driveSubmissionsStatus, setDriveSubmissionsStatus] = useState({}); // path -> prepared (bool)
  const [expandedTests, setExpandedTests] = useState({}); // { [testIdx]: boolean }

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
    handleResetToStart();

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
        if (data.pin) {
          setPin(data.pin);
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

  const handleResetToStart = () => {
    setStep('target');
    setSearchQuery('');
    setSelectedIndex(0);
    setSelectedTarget(null);
    setSelectedStudent(null);
    setPendingServerUrl('');
    setPin('');
    setNewPin('');
    setValidationError('');
    setSubmitResult(null);
    setExpandedTests({});
  };

  const handleGoBack = () => {
    setValidationError('');
    setSearchQuery('');
    setSelectedIndex(0);

    if (step === 'add_server' || step === 'add_server_token' || step === 'add_drive' || step === 'student' || step === 'error') {
      setStep('target');
    } else if (step === 'new_student_id' || step === 'new_student_org') {
      setStep('student');
    } else if (step === 'pin' || step === 'confirm_drive') {
      setStep('student');
    } else if (step === 'new_pin') {
      setStep('pin');
    }
  };

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
      const storedToken = remoteTokens?.[item.target] || '';
      setSelectedTarget({ type: 'remote', target: item.target, token: storedToken });
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
      const nextStep = selectedTarget.type === 'remote' ? 'pin' : 'confirm_drive';
      setStep(nextStep);
      setSearchQuery(nextStep === 'pin' && pin ? pin : '');
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
          setPendingServerUrl(cleanUrl);
          setStep('add_server_token');
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

  const handleAddServerTokenConfirm = (tokenInput) => {
    const token = (tokenInput !== undefined ? tokenInput : searchQuery).trim();
    if (!pendingServerUrl) return;
    addRemoteServer(pendingServerUrl, token);
    setSelectedTarget({ type: 'remote', target: pendingServerUrl, token: token });
    setStep('student');
    setSearchQuery('');
    setSelectedIndex(0);
    setValidationError('');
  };

  const handleFinalSubmit = async (overridePin, overrideNewPin) => {
    setValidationError('');
    setStep('processing');

    const pinToSubmit = (overridePin !== undefined ? overridePin : pin).trim();
    const newPinToSubmit = (overrideNewPin !== undefined ? overrideNewPin : newPin).trim();

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
          pin: pinToSubmit,
          new_pin: newPinToSubmit,
          bearer_token: selectedTarget?.token || ''
        })
      });

      const data = await res.json();
      if (!res.ok) {
        setValidationError(data.error || 'Submission failed');
        setStep('error');
        setSelectedIndex(0);
      } else {
        setSubmitResult(data.result);
        setStep('success');
        setSelectedIndex(0);
      }
    } catch (err) {
      setValidationError('Connection error: ' + err.message);
      setStep('error');
      setSelectedIndex(0);
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

  const getErrorItems = () => {
    return [
      {
        action: 'retry',
        label: '↻ Retry Submission',
        desc: 'Return to initial step to re-select target connection or profile'
      },
      {
        action: 'close',
        label: 'Close',
        desc: 'Exit the submission wizard'
      }
    ];
  };

  const getSuccessItems = () => {
    return [
      {
        action: 'copy',
        label: 'Copy Results',
        desc: 'Copy evaluation summary receipt to clipboard'
      },
      {
        action: 'close',
        label: 'Close',
        desc: 'Exit the submission wizard'
      }
    ];
  };

  const handleCopyPath = () => {
    if (!submitResult) return;
    navigator.clipboard.writeText(submitResult);
    setValidationError('Copied results to clipboard!');
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
      step === 'error' ? getErrorItems() :
      step === 'success' ? getSuccessItems() : [];

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
      } else if (step === 'add_server_token') {
        handleAddServerTokenConfirm();
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
          let currentPin = pin;
          let currentNewPin = newPin;
          if (step === 'pin') {
            const cleanPin = searchQuery.trim() || pin;
            if (!cleanPin) {
              setValidationError('PIN code is required');
              return;
            }
            setPin(cleanPin);
            currentPin = cleanPin;
          } else if (step === 'new_pin') {
            const cleanNewPin = searchQuery.trim() || newPin;
            setNewPin(cleanNewPin);
            currentNewPin = cleanNewPin;
          }
          handleFinalSubmit(currentPin, currentNewPin);
        } else if (actionItem.action === 'change_pin') {
          setStep('new_pin');
          setSearchQuery('');
          setSelectedIndex(0);
          setValidationError('');
        } else if (actionItem.action === 'cancel') {
          onClose();
        }
      } else if (step === 'error' && items[selectedIndex]) {
        const actionItem = items[selectedIndex];
        if (actionItem.action === 'retry') {
          handleResetToStart();
        } else if (actionItem.action === 'close') {
          onClose();
        }
      } else if (step === 'success' && items[selectedIndex]) {
        const actionItem = items[selectedIndex];
        if (actionItem.action === 'copy') {
          handleCopyPath();
        } else if (actionItem.action === 'close') {
          onClose();
        }
      }
    }
  };

  const toggleTestExpanded = (idx) => {
    setExpandedTests(prev => ({
      ...prev,
      [idx]: !prev[idx]
    }));
  };

  // Render parsed evaluation response nicely inside success view
  const renderEvaluationFeedbackView = (resultStr) => {
    if (!resultStr) return null;
    try {
      const parsed = JSON.parse(resultStr);
      if (parsed && parsed.earned_points !== undefined) {
        const totalScore = parsed.max_points > 0 ? Math.round((parsed.earned_points / parsed.max_points) * 100) : 0;
        const isPass = totalScore >= 70;

        return (
          <div style={{ padding: '14px', backgroundColor: 'var(--bg-main)', borderBottom: '1px solid var(--border-color)', display: 'flex', flexDirection: 'column', gap: '10px' }}>
            {/* Score Card Badge */}
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '10px 14px', borderRadius: '6px', backgroundColor: 'var(--bg-card)', border: '1px solid var(--border-color)' }}>
              <div>
                <div style={{ fontSize: '0.7rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                  Evaluation Score
                </div>
                <div style={{ fontSize: '1rem', fontWeight: 700, color: isPass ? 'var(--accent)' : 'var(--accent-red)', marginTop: '2px' }}>
                  {parsed.earned_points} / {parsed.max_points} Points ({totalScore}%)
                </div>
              </div>
              <span style={{ 
                padding: '4px 10px', 
                borderRadius: '12px', 
                fontSize: '0.72rem', 
                fontWeight: 700, 
                textTransform: 'uppercase',
                backgroundColor: isPass ? 'rgba(16, 185, 129, 0.15)' : 'rgba(239, 68, 68, 0.15)', 
                color: isPass ? 'var(--accent)' : 'var(--accent-red)'
              }}>
                {parsed.status || (isPass ? 'PASSED' : 'FAILED')}
              </span>
            </div>

            {/* Test Results Header */}
            <div style={{ fontSize: '0.72rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', marginTop: '4px' }}>
              Test Execution Feedback ({parsed.results ? parsed.results.length : 0} Targets)
            </div>

            {/* Individual Test Cards */}
            {parsed.results && parsed.results.map((tr, idx) => (
              <div key={idx} style={{ borderRadius: '6px', backgroundColor: 'var(--bg-card)', border: '1px solid var(--border-color)', overflow: 'hidden' }}>
                <div 
                  onClick={() => tr.output && toggleTestExpanded(idx)}
                  style={{ padding: '8px 12px', display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.78rem', fontFamily: 'var(--font-mono)', cursor: tr.output ? 'pointer' : 'default' }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <span style={{ color: 'var(--text-primary)', fontWeight: 600 }}>{tr.command}</span>
                    {tr.output && (
                      <span style={{ fontSize: '0.65rem', color: 'var(--text-muted)' }}>
                        {expandedTests[idx] ? '▲ Hide log' : '▼ View log'}
                      </span>
                    )}
                  </div>
                  <span style={{ color: tr.status === 'pass' ? 'var(--accent)' : 'var(--accent-red)', fontWeight: 700 }}>
                    {tr.status === 'pass' ? `PASSED (${tr.points_earned}/${tr.points_possible})` : 'FAILED'}
                  </span>
                </div>

                {/* Expanded Output Panel */}
                {tr.output && expandedTests[idx] && (
                  <div style={{ padding: '8px 12px', borderTop: '1px solid var(--border-color)', backgroundColor: 'var(--bg-main)' }}>
                    <pre style={{ margin: 0, fontFamily: 'var(--font-mono)', fontSize: '0.72rem', color: 'var(--text-secondary)', whiteSpace: 'pre-wrap', wordBreak: 'break-all', maxHeight: '160px', overflowY: 'auto' }}>
                      {tr.output}
                    </pre>
                  </div>
                )}
              </div>
            ))}
          </div>
        );
      }
    } catch {}

    // Fallback for disk paths or non-JSON output
    const isPath = resultStr.includes('/') || resultStr.includes('\\');
    if (isPath) {
      return (
        <div style={{ padding: '14px', backgroundColor: 'var(--bg-main)', borderBottom: '1px solid var(--border-color)', display: 'flex', flexDirection: 'column', gap: '10px' }}>
          <div style={{ fontSize: '0.82rem', color: 'var(--accent)', fontWeight: 600 }}>
            ✓ Submission package successfully saved to drive storage
          </div>
          <div style={{ 
            fontFamily: 'var(--font-mono)', 
            fontSize: '0.75rem', 
            color: 'var(--text-primary)', 
            backgroundColor: 'var(--bg-card)', 
            border: '1px solid var(--border-color)', 
            padding: '8px 12px', 
            borderRadius: '6px',
            wordBreak: 'break-all'
          }}>
            {resultStr}
          </div>
        </div>
      );
    }

    return (
      <div style={{ padding: '14px', backgroundColor: 'var(--bg-main)', borderBottom: '1px solid var(--border-color)', fontSize: '0.8rem', fontFamily: 'var(--font-mono)', whiteSpace: 'pre-wrap', color: 'var(--text-primary)' }}>
        {resultStr}
      </div>
    );
  };

  const showBackButton = step !== 'target' && step !== 'processing' && step !== 'error' && step !== 'success';
  const hideInputRow = step === 'processing' || step === 'error' || step === 'success';

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
            {showBackButton && (
              <button 
                onClick={handleGoBack}
                title="Go Back"
                style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--primary)', display: 'flex', alignItems: 'center', padding: '2px', marginRight: '4px' }}
              >
                <BackIcon size={14} />
              </button>
            )}
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
          <div style={{ padding: '10px 14px', borderBottom: '1px solid var(--border-color)', backgroundColor: 'var(--bg-main)', fontSize: '0.75rem', color: 'var(--text-muted)', wordBreak: 'break-all' }}>
            <span>Submitting as {selectedStudent ? `${selectedStudent.student_id}(org:${selectedStudent.org_id})` : '<id>(org:<org>)'} at {selectedTarget ? selectedTarget.target : '<location>'}</span>
          </div>
        )}

        {/* Input Row / Form Area */}
        {!hideInputRow && (
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

            {step === 'add_server_token' && (
              <input 
                type="password" 
                value={searchQuery} 
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Type Student Bearer Token (optional - press Enter to skip or save)..."
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
          </div>
        )}

        {/* Validation Errors in non-error step */}
        {validationError && step !== 'error' && (
          <div className="validation-alert-error" style={{ margin: '10px 14px 4px 14px', textAlign: 'left' }}>
            {validationError}
          </div>
        )}

        {/* Results List / Details */}
        <div style={{ maxHeight: '380px', overflowY: 'auto', display: 'flex', flexDirection: 'column' }}>
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

          {step === 'add_server_token' && (
            <div 
              className="quick-open-item active"
              onClick={() => handleAddServerTokenConfirm()}
              style={{ borderBottom: '1px solid var(--border-color)', cursor: 'pointer' }}
            >
              <div className="recent-details">
                <span className="recent-name" style={{ fontSize: '0.82rem' }}>
                  {searchQuery.trim() ? 'Save Token: ••••••••' : 'Skip Token (Unauthenticated Connection)'}
                </span>
                <span className="recent-path" style={{ fontSize: '0.7rem', opacity: 0.6 }}>
                  {searchQuery.trim() ? `Save bearer token for ${pendingServerUrl} and proceed` : `No bearer token provided. Press Enter to proceed.`}
                </span>
              </div>
            </div>
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
                    let currentPin = pin;
                    let currentNewPin = newPin;
                    if (step === 'pin') {
                      const cleanPin = searchQuery.trim() || pin;
                      if (!cleanPin) {
                        setValidationError('PIN code is required');
                        return;
                      }
                      setPin(cleanPin);
                      currentPin = cleanPin;
                    } else if (step === 'new_pin') {
                      const cleanNewPin = searchQuery.trim() || newPin;
                      setNewPin(cleanNewPin);
                      currentNewPin = cleanNewPin;
                    }
                    handleFinalSubmit(currentPin, currentNewPin);
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

          {step === 'processing' && (
            <div style={{ padding: '32px 24px', textAlign: 'center', color: 'var(--text-secondary)', fontSize: '0.8rem', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '12px' }}>
              <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" style={{ animation: 'spin 1.5s linear infinite', color: 'var(--primary)' }}>
                <circle cx="12" cy="12" r="10" strokeDasharray="30" strokeDashoffset="10" />
              </svg>
              <span style={{ fontWeight: 600, color: 'var(--text-primary)' }}>Submitting exercise package...</span>
              <span style={{ fontSize: '0.72rem', color: 'var(--text-muted)' }}>Uploading workspace code and executing remote evaluation sandbox targets</span>
            </div>
          )}

          {step === 'error' && (
            <>
              <div style={{ padding: '16px', backgroundColor: 'rgba(239, 68, 68, 0.08)', borderBottom: '1px solid var(--border-color)', display: 'flex', flexDirection: 'column', gap: '6px' }}>
                <div style={{ fontSize: '0.75rem', fontWeight: 700, color: 'var(--accent-red)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                  Submission Failed
                </div>
                <div style={{ fontSize: '0.82rem', color: 'var(--text-primary)', lineHeight: '1.4', wordBreak: 'break-word' }}>
                  {validationError || 'An error occurred while submitting.'}
                </div>
              </div>

              {getErrorItems().map((item, idx) => (
                <div 
                  key={idx}
                  className={`quick-open-item ${selectedIndex === idx ? 'active' : ''}`}
                  onClick={() => {
                    if (item.action === 'retry') {
                      handleResetToStart();
                    } else if (item.action === 'close') {
                      onClose();
                    }
                  }}
                  onMouseEnter={() => setSelectedIndex(idx)}
                  style={{ borderBottom: '1px solid var(--border-color)', cursor: 'pointer' }}
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
            </>
          )}

          {step === 'success' && (
            <>
              {renderEvaluationFeedbackView(submitResult)}

              {getSuccessItems().map((item, idx) => (
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
                  style={{ borderBottom: '1px solid var(--border-color)', cursor: 'pointer' }}
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
            </>
          )}
        </div>
      </div>
    </div>
  );
}
