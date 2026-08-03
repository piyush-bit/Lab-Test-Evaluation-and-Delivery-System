import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useStudent } from '../StudentContext';
import { OpenFolderIcon, CreateIcon } from '../components/Icons';
import RecentWorkspaces from '../components/RecentWorkspaces';
import SystemStatus from '../components/SystemStatus';

export default function Welcome() {
  const navigate = useNavigate();
  const {
    validationError,
    setValidationError,
    targetPath,
    currentCwd,
    triggerQuickOpen,
    triggerOpenWorkspace,
    handleOpenWorkspace,
    triggerExercisePicker,
    handleInitWorkspace
  } = useStudent();

  return (
    <div className="vscode-welcome-container">
      <header className="vscode-header">
        <h1>TDES Console</h1>
        <p>Lab evaluation and packaging dashboard</p>
        {validationError && (
          <div className="validation-alert-error" style={{ marginTop: '14px' }}>
            {validationError}
          </div>
        )}
      </header>

      <div className="vscode-columns-grid">
        {/* LEFT COLUMN: Start & Recents */}
        <div className="vscode-left-col">
          <div className="vscode-section">
            <h2>Start</h2>
            <div className="vscode-action-list">
              <div 
                className="vscode-action-item" 
                onClick={() => {
                  setValidationError('');
                  triggerOpenWorkspace(targetPath || currentCwd, (path) => {
                    handleOpenWorkspace(path);
                  });
                }}
              >
                <span className="action-icon"><OpenFolderIcon /></span>
                <div className="action-details">
                  <span className="action-title">Open Folder...</span>
                  <span className="action-desc">Open an already initialized exercise directory</span>
                </div>
              </div>

              <div 
                className="vscode-action-item" 
                onClick={() => {
                  setValidationError('');
                  triggerExercisePicker((selectedExercise) => {
                    // Step 2: Open directory picker to choose where to unpack it
                    triggerQuickOpen(currentCwd || '~/', (targetFolder) => {
                      const separator = targetFolder.includes('/') ? '/' : '\\';
                      const cleanFolder = targetFolder.endsWith(separator) ? targetFolder : targetFolder + separator;
                      const finalDir = cleanFolder + selectedExercise.lab_id;
                      
                      // Step 3: Run initialization
                      handleInitWorkspace(selectedExercise.lab_id, selectedExercise.version, finalDir);
                    });
                  });
                }}
              >
                <span className="action-icon"><CreateIcon /></span>
                <div className="action-details">
                  <span className="action-title">Initialize Lab...</span>
                  <span className="action-desc">Fetch and initialize a fresh template package</span>
                </div>
              </div>
            </div>
          </div>

          <RecentWorkspaces />
        </div>

        {/* RIGHT COLUMN: Status & Appearance */}
        <SystemStatus />
      </div>
    </div>
  );
}
