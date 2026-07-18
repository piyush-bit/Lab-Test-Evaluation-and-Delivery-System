import React from 'react';
import { useStudent } from '../StudentContext';
import { HistoryIcon } from './Icons';

export default function RecentWorkspaces() {
  const { recents, handleOpenWorkspace } = useStudent();

  return (
    <div className="vscode-section" style={{ marginTop: '28px' }}>
      <h2>Recent</h2>
      <div className="vscode-recents-list">
        {recents.length === 0 ? (
          <div className="empty-recents-msg">No recent folders opened.</div>
        ) : (
          recents.map((item, idx) => {
            const rPath = typeof item === 'string' ? item : item.path;
            const rTitle = typeof item === 'string' ? rPath.split('/').pop() || rPath.split('\\').pop() : item.title;
            const rLabId = typeof item === 'string' ? '' : item.lab_id;
            return (
              <div 
                key={idx} 
                className="vscode-recent-item"
                onClick={() => handleOpenWorkspace(rPath)}
              >
                <span className="recent-icon"><HistoryIcon /></span>
                <div className="recent-details">
                  <span className="recent-name">
                    {rTitle} {rLabId && <span className="recent-lab-id">({rLabId})</span>}
                  </span>
                  <span className="recent-path">{rPath}</span>
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
