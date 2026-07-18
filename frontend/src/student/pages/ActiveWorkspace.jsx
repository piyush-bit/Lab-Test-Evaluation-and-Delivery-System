import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useStudent } from '../StudentContext';
import { BackIcon, CheckIcon } from '../components/Icons';

export default function ActiveWorkspace() {
  const navigate = useNavigate();
  const { activeWorkspacePath, setActiveWorkspacePath } = useStudent();

  // If there's no active workspace, redirect back to welcome
  useEffect(() => {
    if (!activeWorkspacePath) {
      navigate('/');
    }
  }, [activeWorkspacePath, navigate]);

  if (!activeWorkspacePath) return null;

  return (
    <div className="active-workspace-panel glass-card">
      <header className="workspace-header-row">
        <div className="workspace-status-badge">
          <CheckIcon /> Active Workspace
        </div>
        <button 
          className="back-btn" 
          onClick={() => {
            setActiveWorkspacePath('');
            navigate('/');
          }}
        >
          <BackIcon /> Close Workspace
        </button>
      </header>

      <p className="workspace-full-path">
        📁 {activeWorkspacePath}
      </p>

      <div className="placeholder-details-box">
        <h3>Exercise files loaded.</h3>
        <p>
          The workspace is successfully configured. You can edit your source files directly. 
          Ready to run sandbox tests or submit answers.
        </p>
        <div className="next-steps-list">
          <h4>Next Steps:</h4>
          <ul>
            <li>Write your code in your preferred text editor.</li>
            <li>Run <code>euc2 run</code> in your terminal to evaluate public tests.</li>
          </ul>
        </div>
      </div>
    </div>
  );
}
