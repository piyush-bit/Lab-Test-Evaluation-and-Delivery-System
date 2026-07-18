import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useStudent } from '../StudentContext';
import { BackIcon, OpenFolderIcon } from '../components/Icons';

export default function CreateWorkspace() {
  const navigate = useNavigate();
  const {
    validationError,
    setValidationError,
    labID,
    setLabID,
    version,
    setVersion,
    remoteURL,
    setRemoteURL,
    orgID,
    setOrgID,
    targetPath,
    setTargetPath,
    triggerQuickOpen,
    handleCreateWorkspace,
    loading
  } = useStudent();

  return (
    <div className="action-form-container glass-card">
      <button className="back-btn" onClick={() => navigate('/')}>
        <BackIcon /> Back
      </button>
      <h2 className="form-title">Initialize Exercise</h2>
      <p className="form-description">Enter details to retrieve the lab boilerplate and configure files.</p>

      <div className="form-grid">
        <div className="form-group-wrap">
          <label className="field-label">Lab/Exercise ID</label>
          <input 
            type="text" 
            value={labID} 
            onChange={(e) => setLabID(e.target.value)}
            placeholder="e.g. go101-lab01"
            className="text-field"
          />
        </div>

        <div className="form-group-wrap">
          <label className="field-label">Version</label>
          <input 
            type="text" 
            value={version} 
            onChange={(e) => setVersion(e.target.value)}
            placeholder="e.g. v1.0"
            className="text-field"
          />
        </div>

        <div className="form-group-wrap">
          <label className="field-label">Remote Registry URL</label>
          <input 
            type="text" 
            value={remoteURL} 
            onChange={(e) => setRemoteURL(e.target.value)}
            placeholder="e.g. http://localhost:8080"
            className="text-field font-mono"
          />
        </div>

        <div className="form-group-wrap">
          <label className="field-label">Organization/Access ID</label>
          <input 
            type="text" 
            value={orgID} 
            onChange={(e) => setOrgID(e.target.value)}
            placeholder="e.g. default"
            className="text-field"
          />
        </div>

        <div className="form-group-wrap span-2">
          <label className="field-label">Local Working Directory</label>
          <div className="input-with-action-row">
            <input 
              type="text" 
              value={targetPath} 
              onChange={(e) => setTargetPath(e.target.value)}
              placeholder="e.g. ~/Developer/workspace"
              className="text-field font-mono"
            />
            <button 
              className="action-btn"
              onClick={() => {
                triggerQuickOpen(targetPath, (path) => {
                  setTargetPath(path);
                });
              }}
              title="Browse folders"
            >
              <OpenFolderIcon />
            </button>
          </div>
        </div>
      </div>

      {validationError && (
        <div className="validation-alert-error" style={{ marginTop: '16px' }}>
          {validationError}
        </div>
      )}

      <div className="form-actions-bar">
        <button 
          className="btn btn-secondary" 
          onClick={() => navigate('/')}
          disabled={loading}
        >
          Cancel
        </button>
        <button 
          className="btn btn-primary" 
          onClick={handleCreateWorkspace}
          disabled={loading || !labID || !targetPath}
        >
          {loading ? 'Initializing...' : 'Create Workspace'}
        </button>
      </div>
    </div>
  );
}
