import React from 'react';
import { useStudent } from '../StudentContext';

export default function SystemStatus() {
  const { daemonConnected, dockerRunning, isDarkMode, setIsDarkMode } = useStudent();

  return (
    <div className="vscode-right-col">
      <div className="vscode-section">
        <h2>System Status</h2>
        <div className="vscode-technical-status-stack">
          <div className="tech-status-row">
            <span className="tech-label">DAEMON:</span>
            <span className={`tech-value status-${daemonConnected ? 'healthy' : 'dead'}`}>
              {daemonConnected ? 'ONLINE' : 'OFFLINE'}
            </span>
          </div>

          <div className="tech-status-row">
            <span className="tech-label">DOCKER:</span>
            <span className={`tech-value status-${dockerRunning ? 'healthy' : 'warn'}`}>
              {dockerRunning ? 'RUNNING' : 'STOPPED'}
            </span>
          </div>
        </div>
      </div>

      <div className="vscode-section" style={{ marginTop: '28px' }}>
        <h2>Appearance</h2>
        <div className="vscode-technical-status-stack">
          <div className="tech-status-row">
            <span className="tech-label">THEME:</span>
            <button 
              className="vscode-theme-text-btn" 
              onClick={() => setIsDarkMode(!isDarkMode)}
            >
              {isDarkMode ? 'DARK' : 'LIGHT'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
