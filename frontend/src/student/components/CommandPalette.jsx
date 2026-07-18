import React from 'react';
import { useStudent } from '../StudentContext';

export default function CommandPalette() {
  const {
    showQuickOpen,
    quickOpenPath,
    setQuickOpenPath,
    handleQuickOpenKeyDown,
    handleConfirmQuickOpen,
    quickOpenRef,
    quickOpenActiveManifest,
    manifestIdx,
    upIdx,
    dirIndices,
    selectedIndex,
    setSelectedIndex,
    quickOpenParent,
    handleGoUp,
    filteredDirs,
    handleQuickOpenNavigate,
    getPathParts
  } = useStudent();

  if (!showQuickOpen) return null;

  return (
    <div className="quick-open-overlay-blur">
      <div className="quick-open-container" ref={quickOpenRef}>
        <div className="quick-open-input-row">
          <input 
            type="text" 
            value={quickOpenPath} 
            onChange={(e) => setQuickOpenPath(e.target.value)}
            onKeyDown={handleQuickOpenKeyDown}
            placeholder="Search folders..."
            className="quick-open-input"
            autoFocus
          />
          <button className="quick-open-confirm-btn" onClick={handleConfirmQuickOpen}>
            Confirm
          </button>
        </div>

        {/* Parsed Workspace Manifest Preview - placed ABOVE folder names */}
        {quickOpenActiveManifest && (
          <div 
            className={`quick-open-manifest-preview ${selectedIndex === manifestIdx ? 'active' : ''}`}
            onClick={handleConfirmQuickOpen} 
            onMouseEnter={() => setSelectedIndex(manifestIdx)}
            style={{ cursor: 'pointer' }}
            title="Click to open this workspace"
          >
            <div className="manifest-preview-icon">
              <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" className="tdes-logo-svg">
                <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" />
              </svg>
            </div>
            <div className="manifest-preview-text">
              <div className="manifest-preview-id">
                {quickOpenActiveManifest.lab_id}
                <span className="manifest-preview-version">v{quickOpenActiveManifest.version}</span>
              </div>
              <div className="manifest-preview-title">
                {quickOpenActiveManifest.title}
              </div>
            </div>
          </div>
        )}

        <div className="quick-open-results-list">
          {quickOpenParent && (
            <div 
              className={`quick-open-item go-up ${selectedIndex === upIdx ? 'active' : ''}`} 
              onClick={handleGoUp}
              onMouseEnter={() => setSelectedIndex(upIdx)}
            >
              <span className="item-text">..</span>
            </div>
          )}

          {filteredDirs.map((dir, idx) => {
            const globalIdx = dirIndices[idx];
            return (
              <div 
                key={idx}
                className={`quick-open-item ${selectedIndex === globalIdx ? 'active' : ''}`}
                onClick={() => {
                  const { baseDir, sep } = getPathParts(quickOpenPath);
                  handleQuickOpenNavigate(baseDir + dir + sep);
                }}
                onMouseEnter={() => setSelectedIndex(globalIdx)}
              >
                <span className="item-text">{dir}</span>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
