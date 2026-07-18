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
  const [quickOpenPath, setQuickOpenPath] = useState('');
  const [quickOpenDirs, setQuickOpenDirs] = useState([]);
  const [quickOpenParent, setQuickOpenParent] = useState('');
  const [quickOpenCallback, setQuickOpenCallback] = useState(null);
  const [quickOpenActiveManifest, setQuickOpenActiveManifest] = useState(null);
  const [selectedIndex, setSelectedIndex] = useState(0);

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
      }
    }
    function handleEscKey(event) {
      if (event.key === 'Escape') {
        setShowQuickOpen(false);
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

  // Detect baseDir transitions to fetch directories, ignoring typing filtering queries
  useEffect(() => {
    if (!showQuickOpen) return;

    const { baseDir } = getPathParts(quickOpenPath);
    if (baseDir && baseDir !== lastFetchedBaseDirRef.current) {
      lastFetchedBaseDirRef.current = baseDir;
      fetchQuickOpenDirs(baseDir);
    }
  }, [quickOpenPath, showQuickOpen]);

  // Trigger folder picker modal (Quick Open style)
  const triggerQuickOpen = (initialPath, callbackFn) => {
    const startPath = initialPath || currentCwd || '~/';
    setQuickOpenPath(startPath);
    
    const { baseDir } = getPathParts(startPath);
    lastFetchedBaseDirRef.current = baseDir;
    fetchQuickOpenDirs(baseDir);
    
    setQuickOpenCallback(() => callbackFn);
    setShowQuickOpen(true);
  };

  const checkWorkspaceManifest = async (path) => {
    if (!path) {
      setQuickOpenActiveManifest(null);
      return;
    }
    try {
      const res = await fetch(`/api/validate-workspace?path=${encodeURIComponent(path)}`);
      if (res.ok) {
        const data = await res.json();
        if (data.valid && data.manifest) {
          setQuickOpenActiveManifest(data.manifest);
        } else {
          setQuickOpenActiveManifest(null);
        }
      } else {
        setQuickOpenActiveManifest(null);
      }
    } catch {
      setQuickOpenActiveManifest(null);
    }
  };

  // Debounced check of workspace manifest when typing paths
  useEffect(() => {
    if (!showQuickOpen || !quickOpenPath) {
      setQuickOpenActiveManifest(null);
      return;
    }

    const delayDebounce = setTimeout(() => {
      checkWorkspaceManifest(quickOpenPath);
    }, 200);

    return () => clearTimeout(delayDebounce);
  }, [quickOpenPath, showQuickOpen]);

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

  const handleConfirmQuickOpen = () => {
    if (quickOpenCallback) {
      quickOpenCallback(quickOpenPath);
    }
    setShowQuickOpen(false);
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

  // Flatten selectable items list
  const selectableItems = [];
  let manifestIdx = -1;
  let upIdx = -1;
  const dirIndices = [];

  if (quickOpenActiveManifest) {
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

  // Reset focus when elements list changes
  useEffect(() => {
    setSelectedIndex(0);
  }, [filteredDirs.length, quickOpenActiveManifest, quickOpenParent]);

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
        if (item.type === 'manifest') {
          handleConfirmQuickOpen();
        } else if (item.type === 'up') {
          handleGoUp();
        } else if (item.type === 'dir') {
          const { baseDir, sep } = getPathParts(quickOpenPath);
          handleQuickOpenNavigate(baseDir + item.name + sep);
        }
      } else {
        handleConfirmQuickOpen();
      }
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
      handleCreateWorkspace
    }}>
      {children}
    </StudentContext.Provider>
  );
};
