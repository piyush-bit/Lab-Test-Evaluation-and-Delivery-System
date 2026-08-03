import React, { createContext, useContext, useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';

const StudentContext = createContext();

export const useStudent = () => useContext(StudentContext);

export const StudentProvider = ({ children }) => {
  const navigate = useNavigate();

  const [isDarkMode, setIsDarkMode] = useState(true);

  // Server & Daemon Status
  const [daemonConnected, setDaemonConnected] = useState(true);
  const [dockerRunning, setDockerRunning] = useState(false);
  const [currentCwd, setCurrentCwd] = useState('');

  // VS Code-style Command Palette folder picker
  const [showQuickOpen, setShowQuickOpen] = useState(false);
  const [quickOpenSelectedExercise, setQuickOpenSelectedExercise] = useState(null);
  const [quickOpenPath, setQuickOpenPath] = useState('');
  const [quickOpenDirs, setQuickOpenDirs] = useState([]);
  const [quickOpenParent, setQuickOpenParent] = useState('');
  const [quickOpenCallback, setQuickOpenCallback] = useState(null);
  const [quickOpenActiveManifest, setQuickOpenActiveManifest] = useState(null);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [quickOpenMode, setQuickOpenMode] = useState('browse'); // 'browse', 'select_source', 'input_remote', 'input_drive'

  const [recentRemotes, setRecentRemotes] = useState(() => {
    try {
      const stored = localStorage.getItem('recent_remotes');
      return stored ? JSON.parse(stored) : ['http://localhost:8080', 'http://localhost:9090'];
    } catch {
      return ['http://localhost:8080', 'http://localhost:9090'];
    }
  });

  const [remoteServers, setRemoteServers] = useState(() => {
    try {
      const stored = localStorage.getItem('remote_servers');
      return stored ? JSON.parse(stored) : ['http://localhost:8080', 'http://localhost:9090'];
    } catch {
      return ['http://localhost:8080', 'http://localhost:9090'];
    }
  });

  const [remoteServerStatuses, setRemoteServerStatuses] = useState({});

  const normalizeServerUrl = (url) => {
    let cleaned = (url || '').trim();
    if (!cleaned) return '';
    if (!cleaned.startsWith('http://') && !cleaned.startsWith('https://')) {
      cleaned = 'http://' + cleaned;
    }
    return cleaned.replace(/\/+$/, '');
  };

  const checkRemoteServerHealth = async (url) => {
    const normUrl = normalizeServerUrl(url);
    if (!normUrl) return;

    setRemoteServerStatuses(prev => ({
      ...prev,
      [normUrl]: { loading: true, online: prev[normUrl]?.online || false }
    }));

    try {
      const res = await fetch(`/api/remote/health?url=${encodeURIComponent(normUrl)}`);
      if (res.ok) {
        const data = await res.json();
        setRemoteServerStatuses(prev => ({
          ...prev,
          [normUrl]: { loading: false, online: !!data.online, error: data.error || '' }
        }));
      } else {
        setRemoteServerStatuses(prev => ({
          ...prev,
          [normUrl]: { loading: false, online: false, error: 'HTTP error' }
        }));
      }
    } catch (err) {
      setRemoteServerStatuses(prev => ({
        ...prev,
        [normUrl]: { loading: false, online: false, error: err.message }
      }));
    }
  };

  const checkAllRemoteServersHealth = () => {
    remoteServers.forEach(serverUrl => {
      checkRemoteServerHealth(serverUrl);
    });
  };

  const addRemoteServer = (url) => {
    const normUrl = normalizeServerUrl(url);
    if (!normUrl) return;

    setRemoteServers(prev => {
      const filtered = prev.filter(s => normalizeServerUrl(s) !== normUrl);
      const updated = [normUrl, ...filtered];
      localStorage.setItem('remote_servers', JSON.stringify(updated));
      localStorage.setItem('recent_remotes', JSON.stringify(updated));
      setRecentRemotes(updated);
      return updated;
    });

    checkRemoteServerHealth(normUrl);
  };

  const removeRemoteServer = (url) => {
    const normUrl = normalizeServerUrl(url);
    setRemoteServers(prev => {
      const updated = prev.filter(s => normalizeServerUrl(s) !== normUrl);
      localStorage.setItem('remote_servers', JSON.stringify(updated));
      localStorage.setItem('recent_remotes', JSON.stringify(updated));
      setRecentRemotes(updated);
      return updated;
    });

    setRemoteServerStatuses(prev => {
      const copy = { ...prev };
      delete copy[normUrl];
      return copy;
    });
  };

  // Run health check on startup
  useEffect(() => {
    checkAllRemoteServersHealth();
  }, []);

  const [recentDrives, setRecentDrives] = useState(() => {
    try {
      const stored = localStorage.getItem('recent_drives');
      return stored ? JSON.parse(stored) : [];
    } catch {
      return [];
    }
  });

  // Cached exercises list state
  const [quickOpenExercises, setQuickOpenExercises] = useState([]);

  // Fetch Wizard State
  const [fetchSelectedSourceType, setFetchSelectedSourceType] = useState('remote');
  const [fetchSelectedSourceValue, setFetchSelectedSourceValue] = useState('');
  const [fetchInputExerciseID, setFetchInputExerciseID] = useState('');
  const [fetchInputVersion, setFetchInputVersion] = useState('');
  const [driveExercisesList, setDriveExercisesList] = useState([]);
  const [remoteExercisesList, setRemoteExercisesList] = useState([]);

  // Test Runner states
  const [runMode, setRunMode] = useState('docker'); // 'docker' | 'local'
  const [runStatus, setRunStatus] = useState('idle'); // 'idle' | 'running' | 'success' | 'error'
  const [runResults, setRunResults] = useState([]);
  const [runOutput, setRunOutput] = useState('');
  const [showRunPanel, setShowRunPanel] = useState(false);
  const [earnedPoints, setEarnedPoints] = useState(0);
  const [maxPoints, setMaxPoints] = useState(0);

  const addRemoteToRecents = (url) => {
    if (!url) return;
    const updated = [url, ...recentRemotes.filter(r => r !== url)].slice(0, 5);
    setRecentRemotes(updated);
    localStorage.setItem('recent_remotes', JSON.stringify(updated));
  };

  const addDriveToRecents = (path) => {
    if (!path) return;
    const updated = [path, ...recentDrives.filter(d => d !== path)].slice(0, 5);
    setRecentDrives(updated);
    localStorage.setItem('recent_drives', JSON.stringify(updated));
  };

  // Recents Storage
  const [recents, setRecents] = useState(() => {
    try {
      const stored = localStorage.getItem('recent_workspaces');
      return stored ? JSON.parse(stored) : [
        { path: '~/Developer/TDES/workspace/go101-lab01', title: 'Slice-Backed Stack in Go', lab_id: 'go101-lab01' },
        { path: '~/Developer/TDES/workspace/python-basics', title: 'Python Basics', lab_id: 'python-basics' }
      ];
    } catch {
      return [];
    }
  });

  // Create Workspace inputs
  const [labID, setLabID] = useState('go101-lab01');
  const [version, setVersion] = useState('v1.0');
  const [remoteURL, setRemoteURL] = useState('http://localhost:8080');
  const [orgID, setOrgID] = useState('default');
  const [targetPath, setTargetPath] = useState('');

  // Active Workspace State
  const [activeWorkspacePath, setActiveWorkspacePath] = useState('');
  const [validationError, setValidationError] = useState('');
  const [loading, setLoading] = useState(false);

  // Admin Drive State
  const [activeAdminDrivePath, setActiveAdminDrivePath] = useState('');
  const [adminDriveManifest, setAdminDriveManifest] = useState(null);
  const [adminDriveExercises, setAdminDriveExercises] = useState([]);
  const [adminSubmissions, setAdminSubmissions] = useState({ prepared: false, submissions: [] });
  const [adminNotification, setAdminNotification] = useState(null);

  const inspectAdminDrive = async (targetPath) => {
    if (!targetPath) return;
    try {
      const vRes = await fetch(`/api/validate-drive?path=${encodeURIComponent(targetPath)}`);
      if (vRes.ok) {
        const vData = await vRes.json();
        if (vData.valid) {
          setAdminDriveManifest(vData.manifest);
        } else {
          setAdminDriveManifest(null);
        }
      }
      const eRes = await fetch(`/api/drive-exercises?path=${encodeURIComponent(targetPath)}`);
      if (eRes.ok) {
        const eData = await eRes.json();
        const list = [];
        Object.entries(eData).forEach(([exId, entry]) => {
          if (entry && entry.versions) {
            Object.keys(entry.versions).forEach(ver => {
              list.push({ exercise_id: exId, version: ver });
            });
          }
        });
        setAdminDriveExercises(list);
      }
      const sRes = await fetch(`/api/drive/submissions?path=${encodeURIComponent(targetPath)}`);
      if (sRes.ok) {
        const sData = await sRes.json();
        setAdminSubmissions(sData);
      }
    } catch {
      // silent
    }
  };

  const handleOpenAdminDrive = async (path) => {
    if (!path) return;
    setValidationError('');
    setLoading(true);
    try {
      const res = await fetch(`/api/validate-drive?path=${encodeURIComponent(path)}`);
      if (res.ok) {
        const data = await res.json();
        if (data.valid) {
          addDriveToRecents(path);
          setActiveAdminDrivePath(path);
          setAdminDriveManifest(data.manifest);
          await inspectAdminDrive(path);
          setAdminNotification({ type: 'success', title: 'Drive Opened', message: `Opened drive at ${path}` });
        } else {
          setValidationError(`Invalid drive at ${path}: ${data.error || 'Missing manifest.json'}`);
        }
      } else {
        setValidationError('Failed to validate drive path.');
      }
    } catch (err) {
      setValidationError('Connection error: ' + err.message);
    }
    setLoading(false);
  };

  const handlePrepareAdminDrive = async (path) => {
    if (!path) return;
    setValidationError('');
    setLoading(true);
    try {
      const res = await fetch('/api/drive/prepare', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ drive_path: path })
      });
      if (res.ok) {
        const data = await res.json();
        addDriveToRecents(path);
        setActiveAdminDrivePath(path);
        await inspectAdminDrive(path);
        setAdminNotification({ type: 'success', title: 'Drive Prepared', message: data.message || `Prepared drive at ${path}` });
      } else {
        const data = await res.json();
        setValidationError(`Preparation failed: ${data.error}`);
      }
    } catch (err) {
      setValidationError('Connection error: ' + err.message);
    }
    setLoading(false);
  };

  const handlePrepareAdminSubmission = async (path, recipientPublicKey = '') => {
    if (!path) return;
    setValidationError('');
    setLoading(true);
    try {
      const res = await fetch('/api/drive/prepare-submission', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ drive_path: path, recipient_public_key: recipientPublicKey })
      });
      if (res.ok) {
        const data = await res.json();
        await inspectAdminDrive(path);
        setAdminNotification({ type: 'success', title: 'Submissions Enabled', message: data.message || `Submission module prepared at ${path}` });
      } else {
        const data = await res.json();
        setValidationError(`Submission setup failed: ${data.error}`);
      }
    } catch (err) {
      setValidationError('Connection error: ' + err.message);
    }
    setLoading(false);
  };

  const handleAddExerciseToAdminDrive = async (path, exerciseID, versionVal) => {
    if (!path || !exerciseID || !versionVal) return;
    setValidationError('');
    setLoading(true);
    try {
      const res = await fetch('/api/drive/add-exercise', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ drive_path: path, exercise_id: exerciseID, version: versionVal })
      });
      if (res.ok) {
        const data = await res.json();
        await inspectAdminDrive(path);
        setAdminNotification({ type: 'success', title: 'Exercise Added', message: data.message || `Exercise ${exerciseID}@${versionVal} added to drive.` });
      } else {
        const data = await res.json();
        setValidationError(`Add exercise failed: ${data.error}`);
      }
    } catch (err) {
      setValidationError('Connection error: ' + err.message);
    }
    setLoading(false);
  };

  const handleDeleteExerciseFromAdminDrive = async (path, exerciseID, versionVal) => {
    if (!path || !exerciseID) return;
    setValidationError('');
    setLoading(true);
    try {
      const res = await fetch('/api/drive/delete-exercise', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ drive_path: path, exercise_id: exerciseID, version: versionVal })
      });
      if (res.ok) {
        const data = await res.json();
        await inspectAdminDrive(path);
        setAdminNotification({ type: 'success', title: 'Exercise Removed', message: data.message || `Exercise ${exerciseID} removed from drive.` });
      } else {
        const data = await res.json();
        setValidationError(`Delete exercise failed: ${data.error}`);
      }
    } catch (err) {
      setValidationError('Connection error: ' + err.message);
    }
    setLoading(false);
  };

  const handleGenerateKeyPair = async () => {
    try {
      const res = await fetch('/api/drive/generate-keypair', { method: 'POST' });
      if (res.ok) {
        return await res.json();
      }
    } catch {
      // fallback
    }
    return null;
  };

  const handleBatchEvaluateSubmissions = async (path, recipientPrivateKey = '') => {
    if (!path) return null;
    setValidationError('');
    setLoading(true);
    try {
      const res = await fetch('/api/drive/evaluate-batch', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ drive_path: path, recipient_private_key: recipientPrivateKey })
      });
      if (res.ok) {
        const data = await res.json();
        setLoading(false);
        return data.records || [];
      } else {
        const errData = await res.json();
        setValidationError(`Batch evaluation failed: ${errData.error || 'Check private key'}`);
      }
    } catch (err) {
      setValidationError('Connection error: ' + err.message);
    }
    setLoading(false);
    return null;
  };

  const handleSingleEvaluateSubmission = async (path, filename, recipientPrivateKey = '') => {
    if (!path || !filename) return null;
    setValidationError('');
    setLoading(true);
    try {
      const res = await fetch('/api/drive/evaluate-single', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ drive_path: path, filename: filename, recipient_private_key: recipientPrivateKey })
      });
      if (res.ok) {
        const data = await res.json();
        setLoading(false);
        return data.record || null;
      } else {
        const errData = await res.json();
        setValidationError(`Evaluation failed: ${errData.error || 'Check private key'}`);
      }
    } catch (err) {
      setValidationError('Connection error: ' + err.message);
    }
    setLoading(false);
    return null;
  };

  const handleSavePrivateKeyRecord = async (submissionID, publicKey, privateKey) => {
    if (!submissionID || !privateKey) return false;
    try {
      const res = await fetch('/api/drive/save-key', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          submission_id: submissionID,
          public_key: publicKey,
          private_key: privateKey
        })
      });
      if (res.ok) {
        return true;
      }
    } catch {
      // fallback
    }
    return false;
  };

  const triggerAdminDriveFlow = () => {
    setQuickOpenMode('admin_drive_menu');
    setQuickOpenPath('');
    setSelectedIndex(0);
    setQuickOpenCallback(() => (chosenPath) => {});
    setShowQuickOpen(true);
  };

  const quickOpenRef = useRef(null);
  const lastFetchedBaseDirRef = useRef('');

  // Sync Dark/Light Themes
  useEffect(() => {
    const root = document.documentElement;
    if (isDarkMode) {
      root.classList.add('dark');
    } else {
      root.classList.remove('dark');
    }
  }, [isDarkMode]);

  // Daemon Connection Status Heartbeat Check
  useEffect(() => {
    const fetchStatus = async () => {
      try {
        const res = await fetch('/api/status');
        if (res.ok) {
          const data = await res.json();
          setDaemonConnected(true);
          setDockerRunning(data.docker_running);
          if (data.workspace && !currentCwd) {
            setCurrentCwd(data.workspace);
            setTargetPath(data.workspace);
          }
        } else {
          setDaemonConnected(false);
        }
      } catch {
        setDaemonConnected(false);
      }
    };

    fetchStatus();
    const interval = setInterval(fetchStatus, 3000);
    return () => clearInterval(interval);
  }, [currentCwd]);

  // Command Palette Click Outside & ESC listeners
  useEffect(() => {
    function handleClickOutside(event) {
      if (quickOpenRef.current && !quickOpenRef.current.contains(event.target)) {
        setShowQuickOpen(false);
        setQuickOpenSelectedExercise(null);
      }
    }
    function handleEscKey(event) {
      if (event.key === 'Escape') {
        setShowQuickOpen(false);
        setQuickOpenSelectedExercise(null);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    document.addEventListener("keydown", handleEscKey);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("keydown", handleEscKey);
    };
  }, []);

  // Path parts parser for filtering typed queries
  const getPathParts = (path) => {
    const sep = path.includes('/') ? '/' : '\\';
    const lastIdx = path.lastIndexOf(sep);
    if (lastIdx === -1) {
      return { baseDir: path, filterQuery: '', sep };
    }
    const baseDir = path.substring(0, lastIdx + 1); // includes the last separator
    const filterQuery = path.substring(lastIdx + 1);
    return { baseDir, filterQuery, sep };
  };

  // Fetch folders for Command Palette
  const fetchQuickOpenDirs = async (path) => {
    try {
      const res = await fetch(`/api/browse?path=${encodeURIComponent(path)}`);
      if (res.ok) {
        const data = await res.json();
        setQuickOpenDirs(data.directories || []);
        setQuickOpenParent(data.parent_path || '');
      }
    } catch {
      // silent fail
    }
  };

  const [quickOpenValidDriveInfo, setQuickOpenValidDriveInfo] = useState(null);

  const isFolderBrowseMode = quickOpenMode === 'browse' || quickOpenMode === 'open_workspace' || quickOpenMode === 'admin_browse_prepare' || quickOpenMode === 'admin_browse_open';

  // Detect baseDir transitions to fetch directories, ignoring typing filtering queries
  useEffect(() => {
    if (!showQuickOpen || !isFolderBrowseMode) return;

    const { baseDir } = getPathParts(quickOpenPath);
    if (baseDir && baseDir !== lastFetchedBaseDirRef.current) {
      lastFetchedBaseDirRef.current = baseDir;
      fetchQuickOpenDirs(baseDir);
    }
  }, [quickOpenPath, showQuickOpen, quickOpenMode, isFolderBrowseMode]);

  // Live check drive validity when in admin_browse_open mode
  useEffect(() => {
    if (!showQuickOpen || quickOpenMode !== 'admin_browse_open' || !quickOpenPath) {
      setQuickOpenValidDriveInfo(null);
      return;
    }

    const delayDebounce = setTimeout(async () => {
      try {
        const res = await fetch(`/api/validate-drive?path=${encodeURIComponent(quickOpenPath)}`);
        if (res.ok) {
          const data = await res.json();
          if (data.valid) {
            setQuickOpenValidDriveInfo({ valid: true, manifest: data.manifest, path: data.path });
          } else {
            setQuickOpenValidDriveInfo({ valid: false, error: data.error });
          }
        } else {
          setQuickOpenValidDriveInfo({ valid: false, error: 'Directory check failed' });
        }
      } catch (err) {
        setQuickOpenValidDriveInfo({ valid: false, error: err.message });
      }
    }, 150);

    return () => clearTimeout(delayDebounce);
  }, [quickOpenPath, showQuickOpen, quickOpenMode]);

  // Trigger folder picker modal for target directory selection (Category B: Init Target)
  const triggerQuickOpen = (initialPath, callbackFn) => {
    setQuickOpenMode('browse');
    const startPath = initialPath || currentCwd || '~/';
    setQuickOpenPath(startPath);
    
    const { baseDir } = getPathParts(startPath);
    lastFetchedBaseDirRef.current = baseDir;
    fetchQuickOpenDirs(baseDir);
    
    setQuickOpenCallback(() => callbackFn);
    setShowQuickOpen(true);
  };

  // Trigger opening existing exercise workspace modal (Category A: Icon-Based Entity Selection)
  const triggerOpenWorkspace = (initialPath, callbackFn) => {
    setQuickOpenMode('open_workspace');
    const startPath = initialPath || currentCwd || '~/';
    setQuickOpenPath(startPath);
    
    const { baseDir } = getPathParts(startPath);
    lastFetchedBaseDirRef.current = baseDir;
    fetchQuickOpenDirs(baseDir);
    
    setQuickOpenCallback(() => callbackFn || ((chosenPath) => handleOpenWorkspace(chosenPath)));
    setShowQuickOpen(true);
  };

  // Trigger cached exercises list modal
  const triggerExercisePicker = async (callbackFn) => {
    setQuickOpenSelectedExercise(null);
    setQuickOpenPath('');
    setQuickOpenActiveManifest(null);
    setQuickOpenMode('exercises');
    setSelectedIndex(0);

    try {
      const res = await fetch('/api/exercises');
      if (res.ok) {
        const data = await res.json();
        const list = [];
        Object.entries(data).forEach(([labId, entry]) => {
          Object.entries(entry.versions).forEach(([ver, hash]) => {
            list.push({
              lab_id: labId,
              version: ver,
              label: `${labId} (v${ver})`,
              latest: ver === entry.latest
            });
          });
        });
        setQuickOpenExercises(list);
      }
    } catch {
      setQuickOpenExercises([]);
    }

    setQuickOpenCallback(() => callbackFn);
    setShowQuickOpen(true);
  };

  // Trigger cached exercises list modal for deploying to admin drive
  const triggerAdminAddExercisePicker = async (drivePath) => {
    setQuickOpenSelectedExercise(null);
    setQuickOpenPath('');
    setQuickOpenActiveManifest(null);
    setQuickOpenMode('admin_add_exercise');
    setSelectedIndex(0);

    try {
      const res = await fetch('/api/exercises');
      if (res.ok) {
        const data = await res.json();
        const list = [];
        Object.entries(data).forEach(([labId, entry]) => {
          Object.entries(entry.versions).forEach(([ver, hash]) => {
            list.push({
              lab_id: labId,
              version: ver,
              title: entry.title || '',
              language: entry.language || '',
              label: `${labId} (v${ver})`,
              latest: ver === entry.latest
            });
          });
        });
        setQuickOpenExercises(list);
      }
    } catch {
      setQuickOpenExercises([]);
    }

    setQuickOpenCallback(() => async (selectedExercise) => {
      if (selectedExercise && selectedExercise.lab_id && selectedExercise.version) {
        await handleAddExerciseToAdminDrive(drivePath, selectedExercise.lab_id, selectedExercise.version);
      }
    });
    setShowQuickOpen(true);
  };

  // Trigger confirmation modal via Command Palette for clearing evaluation results
  const triggerAdminConfirmClearResults = (onConfirm) => {
    setQuickOpenSelectedExercise(null);
    setQuickOpenPath('');
    setQuickOpenActiveManifest(null);
    setQuickOpenMode('admin_confirm_clear_results');
    setSelectedIndex(0);
    setQuickOpenCallback(() => onConfirm);
    setShowQuickOpen(true);
  };

  const handleClearAdminSubmissions = async (drivePath) => {
    if (!drivePath) return false;
    setValidationError('');
    setLoading(true);
    try {
      const res = await fetch('/api/drive/clear-submissions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ drive_path: drivePath })
      });
      if (res.ok) {
        setAdminNotification('Cleared all submission packages from drive');
        await inspectAdminDrive(drivePath);
        setLoading(false);
        return true;
      } else {
        const errData = await res.json();
        setValidationError(`Clear failed: ${errData.error || 'Check drive path'}`);
      }
    } catch (err) {
      setValidationError('Connection error: ' + err.message);
    }
    setLoading(false);
    return false;
  };

  const triggerAdminConfirmClearSubmissions = (drivePath) => {
    setQuickOpenSelectedExercise(null);
    setQuickOpenPath('');
    setQuickOpenActiveManifest(null);
    setQuickOpenMode('admin_confirm_clear_submissions');
    setSelectedIndex(0);
    setQuickOpenCallback(() => async () => {
      await handleClearAdminSubmissions(drivePath);
    });
    setShowQuickOpen(true);
  };

  // Trigger source repository/drive selection flow
  const triggerInitializeFlow = (callbackFn) => {
    setQuickOpenPath('');
    setQuickOpenActiveManifest(null);
    setQuickOpenMode('select_source');
    setSelectedIndex(0);
    setQuickOpenCallback(() => callbackFn);
    setShowQuickOpen(true);
  };

  const validationReqIdRef = useRef(0);

  // Debounced check of manifest when typing paths in any folder browse mode
  useEffect(() => {
    // Immediately wipe active manifest state synchronously to prevent UI flickering
    setQuickOpenActiveManifest(null);

    if (!showQuickOpen || !quickOpenPath || !isFolderBrowseMode || quickOpenSelectedExercise) {
      return;
    }

    const currentReqId = ++validationReqIdRef.current;

    const delayDebounce = setTimeout(async () => {
      try {
        if (quickOpenMode === 'admin_browse_open') {
          const driveRes = await fetch(`/api/validate-drive?path=${encodeURIComponent(quickOpenPath)}`);
          if (driveRes.ok) {
            const driveData = await driveRes.json();
            if (currentReqId === validationReqIdRef.current && driveData.valid && driveData.manifest) {
              setQuickOpenActiveManifest({ ...driveData.manifest, is_drive: true });
              return;
            }
          }
        } else if (quickOpenMode === 'open_workspace') {
          const wsRes = await fetch(`/api/validate-workspace?path=${encodeURIComponent(quickOpenPath)}`);
          if (wsRes.ok) {
            const wsData = await wsRes.json();
            if (currentReqId === validationReqIdRef.current && wsData.valid && wsData.manifest) {
              setQuickOpenActiveManifest({ ...wsData.manifest, is_workspace: true });
              return;
            }
          }
        }
      } catch {
        // silent fail
      }
    }, 150);

    return () => clearTimeout(delayDebounce);
  }, [quickOpenPath, showQuickOpen, quickOpenMode, isFolderBrowseMode, quickOpenSelectedExercise]);

  // Debounced search of remote exercises when typing queries
  useEffect(() => {
    if (!showQuickOpen || quickOpenMode !== 'remote_exercises') return;

    const delayDebounce = setTimeout(async () => {
      try {
        const res = await fetch(`/api/remote-exercises?remote_url=${encodeURIComponent(fetchSelectedSourceValue)}&org_id=default&q=${encodeURIComponent(quickOpenPath)}`);
        if (res.ok) {
          const data = await res.json();
          const list = (data || []).map(ex => ({
            lab_id: ex.exercise_id,
            version: ex.version,
            title: ex.title || '',
            language: ex.language || '',
            label: `${ex.exercise_id} (v${ex.version}) - ${ex.title || ''}`,
            latest: true
          }));
          setRemoteExercisesList(list);
        }
      } catch {
        // silent fail during typing
      }
    }, 300);

    return () => clearTimeout(delayDebounce);
  }, [quickOpenPath, showQuickOpen, quickOpenMode, fetchSelectedSourceValue]);

  const handleQuickOpenNavigate = (newPath) => {
    setQuickOpenPath(newPath);
    const { baseDir } = getPathParts(newPath);
    lastFetchedBaseDirRef.current = baseDir;
    fetchQuickOpenDirs(baseDir);
    checkWorkspaceManifest(newPath);
  };

  const handleGoUp = () => {
    if (quickOpenParent) {
      const parentPath = quickOpenParent;
      const sep = parentPath.includes('/') ? '/' : '\\';
      const cleanParent = parentPath.endsWith(sep) ? parentPath : parentPath + sep;
      handleQuickOpenNavigate(cleanParent);
    }
  };

  const handleLoadExercisesFromDrive = async (drivePath) => {
    setValidationError('');
    setLoading(true);
    try {
      const res = await fetch(`/api/drive-exercises?path=${encodeURIComponent(drivePath)}`);
      if (res.ok) {
        const data = await res.json();
        const list = [];
        Object.entries(data).forEach(([labId, entry]) => {
          Object.entries(entry.versions).forEach(([ver, hash]) => {
            list.push({
              lab_id: labId,
              version: ver,
              label: `${labId} (v${ver})`,
              latest: ver === entry.latest
            });
          });
        });
        setDriveExercisesList(list);
        setFetchSelectedSourceType('drive');
        setFetchSelectedSourceValue(drivePath);
        setQuickOpenMode('drive_exercises');
        setSelectedIndex(0);
        setQuickOpenPath('');
        setShowQuickOpen(true);
      } else {
        const errData = await res.json();
        setValidationError(`Failed to load drive: ${errData.error || 'Check path'}`);
        setQuickOpenMode('select_source');
        setQuickOpenPath('');
        setShowQuickOpen(true);
      }
    } catch (err) {
      setValidationError("Error reading drive: " + err.message);
      setQuickOpenMode('select_source');
      setQuickOpenPath('');
      setShowQuickOpen(true);
    }
    setLoading(false);
  };

  const handleLoadExercisesFromRemote = async (remoteUrl) => {
    setValidationError('');
    setLoading(true);
    try {
      const res = await fetch(`/api/remote-exercises?remote_url=${encodeURIComponent(remoteUrl)}&org_id=default`);
      if (res.ok) {
        const data = await res.json();
        const list = (data || []).map(ex => ({
          lab_id: ex.exercise_id,
          version: ex.version,
          title: ex.title || '',
          language: ex.language || '',
          label: `${ex.exercise_id} (v${ex.version}) - ${ex.title || ''}`,
          latest: true
        }));
        setRemoteExercisesList(list);
        setFetchSelectedSourceType('remote');
        setFetchSelectedSourceValue(remoteUrl);
        setQuickOpenMode('remote_exercises');
        setSelectedIndex(0);
        setQuickOpenPath('');
        setShowQuickOpen(true);
      } else {
        const errData = await res.json();
        setValidationError(`Failed to load remote exercises: ${errData.error || 'Check registry URL'}`);
        setQuickOpenMode('select_source');
        setQuickOpenPath('');
        setShowQuickOpen(true);
      }
    } catch (err) {
      setValidationError("Error reading remote registry: " + err.message);
      setQuickOpenMode('select_source');
      setQuickOpenPath('');
      setShowQuickOpen(true);
    }
    setLoading(false);
  };

  const handleExecuteFetchInline = async (selectedExercise, sourceType, sourceValue) => {
    const exerciseId = selectedExercise.lab_id;
    const versionVal = selectedExercise.version;
    setValidationError('');
    setLoading(true);

    try {
      const body = {
        exercise_id: exerciseId,
        version: versionVal,
        org_id: 'default'
      };
      if (sourceType === 'remote') {
        body.remote_url = sourceValue;
      } else {
        body.drive_path = sourceValue;
      }

      const res = await fetch('/api/fetch', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      });

      if (!res.ok) {
        const data = await res.json();
        setValidationError(data.error || 'Fetch failed');
        setLoading(false);
        return;
      }

      // Success: automatically select this exercise and complete the selection flow
      setQuickOpenSelectedExercise(selectedExercise);
      setShowQuickOpen(false);
      if (quickOpenCallback) {
        quickOpenCallback(selectedExercise);
      }
    } catch (err) {
      setValidationError("Connection error: " + err.message);
    }
    setLoading(false);
  };

  const handleConfirmQuickOpen = () => {
    if (quickOpenMode === 'admin_drive_menu') {
      const item = selectableItems[selectedIndex];
      if (item) {
        if (item.type === 'action') {
          if (item.action === 'admin_prepare_drive') {
            setValidationError('');
            setQuickOpenMode('admin_browse_prepare');
            const startPath = currentCwd || '~/';
            setQuickOpenPath(startPath);
            setSelectedIndex(0);
            const { baseDir } = getPathParts(startPath);
            lastFetchedBaseDirRef.current = baseDir;
            fetchQuickOpenDirs(baseDir);
          } else if (item.action === 'admin_open_drive') {
            setValidationError('');
            setQuickOpenMode('admin_browse_open');
            const startPath = currentCwd || '~/';
            setQuickOpenPath(startPath);
            setSelectedIndex(0);
            const { baseDir } = getPathParts(startPath);
            lastFetchedBaseDirRef.current = baseDir;
            fetchQuickOpenDirs(baseDir);
          }
        } else if (item.type === 'recent_drive_disk') {
          setValidationError('');
          setShowQuickOpen(false);
          handleOpenAdminDrive(item.value);
        }
      }
      return;
    } else if (quickOpenMode === 'open_workspace') {
      const item = selectableItems[selectedIndex];
      if (item && item.type === 'manifest' && item.data?.is_workspace) {
        setValidationError('');
        setShowQuickOpen(false);
        if (quickOpenCallback) {
          quickOpenCallback(quickOpenPath);
        } else {
          handleOpenWorkspace(quickOpenPath);
        }
        return;
      }
      return;
    } else if (quickOpenMode === 'admin_browse_prepare') {
      const item = selectableItems[selectedIndex];
      if (item && item.type === 'select_current_dir') {
        setValidationError('');
        setShowQuickOpen(false);
        handlePrepareAdminDrive(item.path);
        return;
      }
    } else if (quickOpenMode === 'admin_browse_open') {
      const item = selectableItems[selectedIndex];
      if (item && item.type === 'manifest' && item.data?.is_drive) {
        setValidationError('');
        setShowQuickOpen(false);
        handleOpenAdminDrive(quickOpenPath);
        return;
      }
      return;
    }

    if (!quickOpenCallback) {
      setShowQuickOpen(false);
      return;
    }

    const currentSelectedItem = selectableItems[selectedIndex];
    if (currentSelectedItem && currentSelectedItem.type === 'select_current_dir') {
      setShowQuickOpen(false);
      quickOpenCallback(quickOpenPath);
      return;
    }

    if (quickOpenMode === 'exercises') {
      const item = selectableItems[selectedIndex];
      if (item) {
        if (item.type === 'action' && item.action === 'fetch_new') {
          setValidationError('');
          setQuickOpenMode('select_source');
          setQuickOpenPath('');
          setSelectedIndex(0);
        } else if (item.type === 'exercise') {
          setQuickOpenSelectedExercise(item.data);
          setShowQuickOpen(false);
          quickOpenCallback(item.data);
        }
      }
    } else if (quickOpenMode === 'admin_add_exercise') {
      const item = selectableItems[selectedIndex];
      if (item && item.type === 'exercise') {
        setShowQuickOpen(false);
        if (quickOpenCallback) {
          quickOpenCallback(item.data);
        }
      }
    } else if (quickOpenMode === 'admin_confirm_clear_results' || quickOpenMode === 'admin_confirm_clear_submissions') {
      const item = selectableItems[selectedIndex];
      setShowQuickOpen(false);
      if (item && item.action === 'yes' && quickOpenCallback) {
        quickOpenCallback();
      }
    } else if (quickOpenMode === 'drive_exercises') {
      const item = selectableItems[selectedIndex];
      if (item && item.type === 'exercise') {
        handleExecuteFetchInline(item.data, 'drive', fetchSelectedSourceValue);
      }
    } else if (quickOpenMode === 'remote_exercises') {
      const item = selectableItems[selectedIndex];
      if (item && item.type === 'exercise') {
        handleExecuteFetchInline(item.data, 'remote', fetchSelectedSourceValue);
      }
    } else if (quickOpenMode === 'select_source') {
      const item = selectableItems[selectedIndex];
      if (item) {
        if (item.type === 'action') {
          if (item.action === 'new_remote') {
            setQuickOpenMode('input_remote');
            setQuickOpenPath('');
            setSelectedIndex(0);
          } else if (item.action === 'new_drive') {
            setValidationError('');
            triggerQuickOpen(currentCwd || '~/', async (chosenPath) => {
              setLoading(true);
              setValidationError('');
              try {
                const res = await fetch(`/api/validate-drive?path=${encodeURIComponent(chosenPath)}`);
                if (res.ok) {
                  const data = await res.json();
                  if (data.valid) {
                    addDriveToRecents(chosenPath);
                    await handleLoadExercisesFromDrive(chosenPath);
                  } else {
                    setValidationError(`Invalid drive manifest at "${chosenPath}": ${data.error || 'Check manifest.json'}`);
                    setQuickOpenMode('select_source');
                    setQuickOpenPath('');
                    setShowQuickOpen(true);
                  }
                } else {
                  setValidationError("Failed to validate drive folder");
                  setQuickOpenMode('select_source');
                  setQuickOpenPath('');
                  setShowQuickOpen(true);
                }
              } catch (err) {
                setValidationError("Connection error: " + err.message);
                setQuickOpenMode('select_source');
                setQuickOpenPath('');
                setShowQuickOpen(true);
              }
              setLoading(false);
            });
          }
        } else if (item.type === 'recent_source') {
          setValidationError('');
          if (item.sourceType === 'drive') {
            handleLoadExercisesFromDrive(item.value);
          } else {
            handleLoadExercisesFromRemote(item.value);
          }
        }
      }
    } else if (quickOpenMode === 'input_remote') {
      if (quickOpenPath.trim()) {
        const val = quickOpenPath.trim();
        addRemoteToRecents(val);
        handleLoadExercisesFromRemote(val);
      }
    } else if (quickOpenMode === 'input_drive') {
      if (quickOpenPath.trim()) {
        const val = quickOpenPath.trim();
        addDriveToRecents(val);
        setValidationError('');
        setFetchSelectedSourceType('drive');
        setFetchSelectedSourceValue(val);
        setQuickOpenMode('input_exercise_id');
        setQuickOpenPath('');
        setSelectedIndex(0);
      }
    } else if (quickOpenMode === 'input_exercise_id') {
      if (quickOpenPath.trim()) {
        const val = quickOpenPath.trim();
        setValidationError('');
        setFetchInputExerciseID(val);
        setQuickOpenMode('input_exercise_version');
        setQuickOpenPath('');
        setSelectedIndex(0);
      }
    } else if (quickOpenMode === 'input_exercise_version') {
      const ver = quickOpenPath.trim();
      const manualEx = {
        lab_id: fetchInputExerciseID,
        version: ver,
        label: `${fetchInputExerciseID} (v${ver})`,
        latest: true
      };
      handleExecuteFetchInline(manualEx, fetchSelectedSourceType, fetchSelectedSourceValue);
    } else {
      setShowQuickOpen(false);
      quickOpenCallback(quickOpenPath);
    }
  };

  const addWorkspaceToRecents = (path, manifest) => {
    const title = (manifest && manifest.title) || path.split('/').pop() || path.split('\\').pop();
    const lab_id = (manifest && manifest.lab_id) || '';
    const newEntry = { path, title, lab_id };
    
    const filtered = recents.filter(r => {
      const rPath = typeof r === 'string' ? r : r.path;
      return rPath !== path;
    });
    
    const updated = [newEntry, ...filtered].slice(0, 5);
    setRecents(updated);
    localStorage.setItem('recent_workspaces', JSON.stringify(updated));
  };

  const removeWorkspaceFromRecents = (pathToRem) => {
    if (!pathToRem) return;
    const normTarget = pathToRem.replace(/[/\\]+$/, '');
    setRecents(prevRecents => {
      const updated = prevRecents.filter(r => {
        const rPath = typeof r === 'string' ? r : r.path;
        const normR = (rPath || '').replace(/[/\\]+$/, '');
        return normR !== normTarget;
      });
      localStorage.setItem('recent_workspaces', JSON.stringify(updated));
      return updated;
    });
  };

  // Open Workspace Action
  const handleOpenWorkspace = async (path) => {
    setValidationError('');
    setLoading(true);

    try {
      const res = await fetch(`/api/validate-workspace?path=${encodeURIComponent(path)}`);
      if (res.ok) {
        const data = await res.json();
        if (data.valid) {
          setActiveWorkspacePath(data.path);
          addWorkspaceToRecents(data.path, data.manifest);
          navigate('/workspace');
        } else {
          setValidationError(data.error || "No manifest.json found. Ensure this is an initialized TDES directory.");
          removeWorkspaceFromRecents(path);
        }
      } else {
        setValidationError("Failed to communicate with local agent.");
      }
    } catch (err) {
      setValidationError("Connection error: " + err.message);
    }
    setLoading(false);
  };

  // Create Workspace Action
  const handleCreateWorkspace = async () => {
    setValidationError('');
    setLoading(true);

    try {
      const fetchRes = await fetch('/api/fetch', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          exercise_id: labID,
          version: version,
          remote_url: remoteURL,
          org_id: orgID
        })
      });
      
      if (!fetchRes.ok) {
        const fetchErr = await fetchRes.json();
        setValidationError(`Fetch failed: ${fetchErr.error || 'Server error'}`);
        setLoading(false);
        return;
      }

      const initRes = await fetch('/api/init', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          exercise_id: labID,
          version: version,
          target_dir: targetPath
        })
      });

      if (!initRes.ok) {
        const initErr = await initRes.json();
        setValidationError(`Init failed: ${initErr.error || 'Server error'}`);
        setLoading(false);
        return;
      }

      // Check targetPath validation to retrieve the parsed manifest details
      const valRes = await fetch(`/api/validate-workspace?path=${encodeURIComponent(targetPath)}`);
      let manifest = null;
      if (valRes.ok) {
        const valData = await valRes.json();
        if (valData.valid) {
          manifest = valData.manifest;
        }
      }

      setActiveWorkspacePath(targetPath);
      addWorkspaceToRecents(targetPath, manifest);
      navigate('/workspace');
    } catch (err) {
      setValidationError("Connection error: " + err.message);
    }
    setLoading(false);
  };

  // Filter visible subdirectories on typed text
  const { filterQuery } = getPathParts(quickOpenPath);
  const filteredDirs = quickOpenDirs.filter(dir => 
    dir.toLowerCase().includes(filterQuery.toLowerCase())
  );

  // Flatten selectable items list based on active mode
  const selectableItems = [];
  let manifestIdx = -1;
  let upIdx = -1;
  const dirIndices = [];

  if (quickOpenMode === 'admin_drive_menu') {
    const q = quickOpenPath.toLowerCase();
    if (!q || "prepare a drive".includes(q) || "prepare drive".includes(q)) {
      selectableItems.push({
        type: 'action',
        action: 'admin_prepare_drive',
        label: 'Prepare a Drive...'
      });
    }
    if (!q || "open drive".includes(q)) {
      selectableItems.push({
        type: 'action',
        action: 'admin_open_drive',
        label: 'Open Drive...'
      });
    }
    recentDrives.forEach(path => {
      if (!q || path.toLowerCase().includes(q)) {
        selectableItems.push({
          type: 'recent_drive_disk',
          value: path,
          label: path
        });
      }
    });
  } else if (quickOpenMode === 'exercises') {
    selectableItems.push({
      type: 'action',
      action: 'fetch_new',
      label: '+ Fetch new exercise from Remote/Drive...'
    });

    const filteredExercises = quickOpenExercises.filter(ex => 
      ex.label.toLowerCase().includes(quickOpenPath.toLowerCase())
    );
    filteredExercises.forEach((ex) => {
      selectableItems.push({ type: 'exercise', data: ex });
    });
  } else if (quickOpenMode === 'admin_add_exercise') {
    const filteredExercises = quickOpenExercises.filter(ex => 
      ex.label.toLowerCase().includes(quickOpenPath.toLowerCase()) ||
      ex.lab_id.toLowerCase().includes(quickOpenPath.toLowerCase())
    );
    filteredExercises.forEach((ex) => {
      selectableItems.push({ type: 'exercise', data: ex });
    });
  } else if (quickOpenMode === 'admin_confirm_clear_results') {
    selectableItems.push({ type: 'action', action: 'yes', label: 'Yes, clear all evaluation results' });
    selectableItems.push({ type: 'action', action: 'no', label: 'No, cancel' });
  } else if (quickOpenMode === 'admin_confirm_clear_submissions') {
    selectableItems.push({ type: 'action', action: 'yes', label: 'Yes, permanently delete all submission packages from drive' });
    selectableItems.push({ type: 'action', action: 'no', label: 'No, cancel' });
  } else if (quickOpenMode === 'select_source') {
    const q = quickOpenPath.toLowerCase();
    if (!q || "+ New Remote Registry URL...".toLowerCase().includes(q)) {
      selectableItems.push({ type: 'action', action: 'new_remote', label: '+ New Remote Registry URL...' });
    }
    if (!q || "+ New Drive Path...".toLowerCase().includes(q)) {
      selectableItems.push({ type: 'action', action: 'new_drive', label: '+ New Drive Path...' });
    }
    recentRemotes.forEach(url => {
      if (!q || url.toLowerCase().includes(q)) {
        selectableItems.push({ type: 'recent_source', sourceType: 'remote', value: url, label: url });
      }
    });
    recentDrives.forEach(path => {
      if (!q || path.toLowerCase().includes(q)) {
        selectableItems.push({ type: 'recent_source', sourceType: 'drive', value: path, label: path });
      }
    });
  } else if (quickOpenMode === 'drive_exercises') {
    const filteredExercises = driveExercisesList.filter(ex => 
      ex.label.toLowerCase().includes(quickOpenPath.toLowerCase())
    );
    filteredExercises.forEach((ex) => {
      selectableItems.push({ type: 'exercise', data: ex });
    });
  } else if (quickOpenMode === 'remote_exercises') {
    remoteExercisesList.forEach((ex) => {
      selectableItems.push({ type: 'exercise', data: ex });
    });
  } else if (quickOpenMode === 'input_remote') {
    selectableItems.push({ type: 'input_confirm', label: quickOpenPath.trim() ? `Connect Registry: ${quickOpenPath}` : 'Type Registry URL and press Enter...' });
  } else if (quickOpenMode === 'input_drive') {
    selectableItems.push({ type: 'input_confirm', label: quickOpenPath.trim() ? `Connect Drive Folder: ${quickOpenPath}` : 'Type Drive folder path and press Enter...' });
  } else if (quickOpenMode === 'input_exercise_id') {
    selectableItems.push({ type: 'input_confirm', label: quickOpenPath.trim() ? `Confirm Exercise ID: ${quickOpenPath}` : 'Type Exercise ID (e.g. go101-lab01) and press Enter...' });
  } else if (quickOpenMode === 'input_exercise_version') {
    selectableItems.push({ type: 'input_confirm', label: quickOpenPath.trim() ? `Confirm Version: ${quickOpenPath}` : 'Type Version (or press Enter for Latest)...' });
  } else if (quickOpenMode === 'open_workspace') {
    // Category A: Icon-Based Entity Selection (Open Existing Exercise Workspace)
    if (quickOpenActiveManifest?.is_workspace) {
      manifestIdx = selectableItems.length;
      selectableItems.push({ type: 'manifest', data: quickOpenActiveManifest });
    }
    if (quickOpenParent) {
      upIdx = selectableItems.length;
      selectableItems.push({ type: 'up' });
    }
    filteredDirs.forEach((dir) => {
      dirIndices.push(selectableItems.length);
      selectableItems.push({ type: 'dir', name: dir });
    });
  } else if (quickOpenMode === 'admin_browse_prepare') {
    // Category B: Folder-Based Target Selection (Prepare Drive Target)
    if (quickOpenPath) {
      const folderName = quickOpenPath.replace(/[/\\]+$/, '').split(/[/\\]/).pop() || quickOpenPath;
      selectableItems.push({
        type: 'select_current_dir',
        path: quickOpenPath,
        label: `Prepare Drive: ${folderName}`
      });
    }
    if (quickOpenParent) {
      upIdx = selectableItems.length;
      selectableItems.push({ type: 'up' });
    }
    filteredDirs.forEach((dir) => {
      dirIndices.push(selectableItems.length);
      selectableItems.push({ type: 'dir', name: dir });
    });
  } else if (quickOpenMode === 'admin_browse_open') {
    // Category A: Icon-Based Entity Selection (Open Existing TDES Drive)
    if (quickOpenActiveManifest?.is_drive) {
      manifestIdx = selectableItems.length;
      selectableItems.push({ type: 'manifest', data: quickOpenActiveManifest });
    }
    if (quickOpenParent) {
      upIdx = selectableItems.length;
      selectableItems.push({ type: 'up' });
    }
    filteredDirs.forEach((dir) => {
      dirIndices.push(selectableItems.length);
      selectableItems.push({ type: 'dir', name: dir });
    });
  } else {
    // Category B: Folder-Based Target Selection (Choose Init Target Directory)
    if (quickOpenPath) {
      const folderName = quickOpenPath.replace(/[/\\]+$/, '').split(/[/\\]/).pop() || quickOpenPath;
      selectableItems.push({
        type: 'select_current_dir',
        path: quickOpenPath,
        label: `Select directory: ${folderName}`
      });
    }
    if (quickOpenParent) {
      upIdx = selectableItems.length;
      selectableItems.push({ type: 'up' });
    }
    filteredDirs.forEach((dir) => {
      dirIndices.push(selectableItems.length);
      selectableItems.push({ type: 'dir', name: dir });
    });
  }

  // Reset focus when elements list changes
  useEffect(() => {
    setSelectedIndex(0);
  }, [filteredDirs.length, quickOpenActiveManifest, quickOpenParent, recentRemotes.length, recentDrives.length, driveExercisesList.length, remoteExercisesList.length, quickOpenMode, quickOpenPath]);

  // Keyboard navigation inside Command Palette input
  const handleQuickOpenKeyDown = (e) => {
    if (selectableItems.length === 0) {
      if (e.key === 'Enter') {
        handleConfirmQuickOpen();
      }
      return;
    }

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelectedIndex(prev => (prev + 1) % selectableItems.length);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelectedIndex(prev => (prev - 1 + selectableItems.length) % selectableItems.length);
    } else if (e.key === 'Enter') {
      e.preventDefault();
      const item = selectableItems[selectedIndex];
      if (item) {
        if (quickOpenMode === 'exercises' || quickOpenMode === 'admin_add_exercise' || quickOpenMode === 'admin_confirm_clear_results' || quickOpenMode === 'admin_confirm_clear_submissions' || quickOpenMode === 'drive_exercises' || quickOpenMode === 'remote_exercises') {
          if (item.type === 'action' && item.action === 'fetch_new') {
            setValidationError('');
            setQuickOpenMode('select_source');
            setQuickOpenPath('');
            setSelectedIndex(0);
          } else if (item.type === 'exercise') {
            handleConfirmQuickOpen();
          }
        } else if (quickOpenMode === 'select_source') {
          if (item.type === 'action') {
            if (item.action === 'new_remote') {
              setQuickOpenMode('input_remote');
              setQuickOpenPath('');
              setSelectedIndex(0);
            } else if (item.action === 'new_drive') {
              handleConfirmQuickOpen();
            }
          } else if (item.type === 'recent_source') {
            handleConfirmQuickOpen();
          }
        } else if (quickOpenMode === 'input_remote' || quickOpenMode === 'input_drive' || quickOpenMode === 'input_exercise_id' || quickOpenMode === 'input_exercise_version') {
          handleConfirmQuickOpen();
        } else {
          if (item.type === 'manifest') {
            handleConfirmQuickOpen();
          } else if (item.type === 'select_current_dir') {
            handleConfirmQuickOpen();
          } else if (item.type === 'up') {
            handleGoUp();
          } else if (item.type === 'dir') {
            const { baseDir, sep } = getPathParts(quickOpenPath);
            handleQuickOpenNavigate(baseDir + item.name + sep);
          }
        }
      } else {
        handleConfirmQuickOpen();
      }
    }
  };

  // Fetch Exercise from Remote/Drive
  const handleExecuteFetch = async () => {
    setValidationError('');
    setLoading(true);

    try {
      const body = {
        exercise_id: fetchLabID,
        version: fetchVersion,
        org_id: fetchOrgID
      };
      if (fetchSourceType === 'remote') {
        body.remote_url = fetchRemoteURL;
      } else {
        body.drive_path = fetchDrivePath;
      }

      const res = await fetch('/api/fetch', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      });

      if (!res.ok) {
        const data = await res.json();
        setValidationError(data.error || 'Fetch failed');
        setLoading(false);
        return;
      }

      // Success: refresh cached list and go back to selection screen
      await triggerExercisePicker(quickOpenCallback);
    } catch (err) {
      setValidationError("Connection error: " + err.message);
    }
    setLoading(false);
  };

  // Initialize Workspace from selected Cached Exercise
  const handleInitWorkspace = async (labId, ver, targetDir) => {
    setValidationError('');
    setLoading(true);

    try {
      const initRes = await fetch('/api/init', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          exercise_id: labId,
          version: ver,
          target_dir: targetDir
        })
      });

      if (!initRes.ok) {
        const initErr = await initRes.json();
        setValidationError(`Initialization failed: ${initErr.error || 'Server error'}`);
        setLoading(false);
        return;
      }

      // Check targetPath validation to retrieve the parsed manifest details
      const valRes = await fetch(`/api/validate-workspace?path=${encodeURIComponent(targetDir)}`);
      let manifest = null;
      if (valRes.ok) {
        const valData = await valRes.json();
        if (valData.valid) {
          manifest = valData.manifest;
        }
      }

      setActiveWorkspacePath(targetDir);
      addWorkspaceToRecents(targetDir, manifest);
      setQuickOpenSelectedExercise(null);
      navigate('/workspace');
    } catch (err) {
      setValidationError("Connection error: " + err.message);
    }
    setLoading(false);
  };

  const handleRunTests = async (specificCommand = '', modeOverride = null) => {
    if (!activeWorkspacePath) return;
    const modeToUse = modeOverride || runMode;
    
    setShowRunPanel(true);
    if (specificCommand) {
      setRunResults(prev => prev.map(item => 
        item.command === specificCommand ? { ...item, status: 'running', output: 'Executing specific command...' } : item
      ));
    } else {
      setRunStatus('running');
      setRunResults([]);
      setRunOutput('');
    }

    try {
      const res = await fetch('/api/workspace/run', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          path: activeWorkspacePath,
          mode: modeToUse,
          command: specificCommand
        })
      });

      if (res.ok) {
        const data = await res.json();
        if (specificCommand) {
          setRunResults(prev => prev.map(item => {
            const newItem = (data.results || []).find(r => r.command === item.command);
            return newItem && newItem.status !== 'idle' ? newItem : item;
          }));
        } else {
          setEarnedPoints(data.earned_points || 0);
          setMaxPoints(data.max_points || 0);
          setRunResults(data.results || []);
          if (data.success) {
            setRunStatus('success');
          } else {
            setRunStatus('error');
          }
        }
      } else {
        const errData = await res.json();
        setRunOutput(errData.error || 'Failed to execute run');
        setRunStatus('error');
      }
    } catch (err) {
      setRunOutput("Connection error: " + err.message);
      setRunStatus('error');
    }
  };

  return (
    <StudentContext.Provider value={{
      isDarkMode, setIsDarkMode,
      daemonConnected, setDaemonConnected,
      dockerRunning, setDockerRunning,
      currentCwd, setCurrentCwd,
      showQuickOpen, setShowQuickOpen,
      quickOpenPath, setQuickOpenPath,
      quickOpenDirs, setQuickOpenDirs,
      quickOpenParent, setQuickOpenParent,
      quickOpenActiveManifest, setQuickOpenActiveManifest,
      selectedIndex, setSelectedIndex,
      recents, setRecents,
      labID, setLabID,
      version, setVersion,
      remoteURL, setRemoteURL,
      orgID, setOrgID,
      targetPath, setTargetPath,
      activeWorkspacePath, setActiveWorkspacePath,
      validationError, setValidationError,
      loading, setLoading,
      quickOpenRef,
      getPathParts,
      triggerQuickOpen,
      triggerOpenWorkspace,
      triggerInitializeFlow,
      handleQuickOpenNavigate,
      handleGoUp,
      handleConfirmQuickOpen,
      handleQuickOpenKeyDown,
      selectableItems,
      manifestIdx,
      upIdx,
      dirIndices,
      filteredDirs,
      handleOpenWorkspace,
      removeWorkspaceFromRecents,
      handleCreateWorkspace,
      handleInitWorkspace,
      quickOpenMode,
      setQuickOpenMode,
      recentRemotes,
      recentDrives,
      triggerExercisePicker,
      quickOpenExercises,
      driveExercisesList,
      handleLoadExercisesFromDrive,
      remoteExercisesList,
      handleLoadExercisesFromRemote,
      quickOpenSelectedExercise,
      setQuickOpenSelectedExercise,
      runMode,
      setRunMode,
      runStatus,
      setRunStatus,
      runResults,
      setRunResults,
      runOutput,
      setRunOutput,
      showRunPanel,
      setShowRunPanel,
      handleRunTests,
      earnedPoints,
      setEarnedPoints,
      maxPoints,
      setMaxPoints,
      activeAdminDrivePath,
      setActiveAdminDrivePath,
      adminDriveManifest,
      setAdminDriveManifest,
      adminDriveExercises,
      setAdminDriveExercises,
      adminSubmissions,
      setAdminSubmissions,
      adminNotification,
      setAdminNotification,
      triggerAdminDriveFlow,
      handleOpenAdminDrive,
      handlePrepareAdminDrive,
      handlePrepareAdminSubmission,
      handleAddExerciseToAdminDrive,
      triggerAdminAddExercisePicker,
      triggerAdminConfirmClearResults,
      handleClearAdminSubmissions,
      triggerAdminConfirmClearSubmissions,
      handleDeleteExerciseFromAdminDrive,
      handleGenerateKeyPair,
      handleBatchEvaluateSubmissions,
      handleSingleEvaluateSubmission,
      handleSavePrivateKeyRecord,
      remoteServers,
      setRemoteServers,
      remoteServerStatuses,
      addRemoteServer,
      removeRemoteServer,
      checkAllRemoteServersHealth,
      checkRemoteServerHealth
    }}>
      {children}
    </StudentContext.Provider>
  );
};
