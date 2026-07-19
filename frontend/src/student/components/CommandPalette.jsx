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
    getPathParts,
    quickOpenMode,
    setQuickOpenMode,
    selectableItems,
    validationError,
    setValidationError,
    loading
  } = useStudent();

  if (!showQuickOpen) return null;

  const isBrowse = quickOpenMode === 'browse';
  const isExercises = quickOpenMode === 'exercises';
  const isSelectSource = quickOpenMode === 'select_source';
  const isInputRemote = quickOpenMode === 'input_remote';
  const isInputDrive = quickOpenMode === 'input_drive';
  const isInputExerciseID = quickOpenMode === 'input_exercise_id';
  const isInputExerciseVersion = quickOpenMode === 'input_exercise_version';
  const isDriveExercises = quickOpenMode === 'drive_exercises';
  const isRemoteExercises = quickOpenMode === 'remote_exercises';

  // Determine placeholder and header labels based on current mode
  let placeholder = "Search folders...";
  let headerLabel = "";

  if (isExercises) {
    placeholder = "Search cached exercises...";
    headerLabel = "SELECT EXERCISE TO INITIALIZE";
  } else if (isSelectSource) {
    placeholder = "Search or select repository/drive location...";
    headerLabel = "SELECT REPOSITORY SOURCE OR DRIVE LOCATION";
  } else if (isInputRemote) {
    placeholder = "Enter remote registry URL (e.g. http://localhost:8080)...";
    headerLabel = "TYPE REGISTRY URL AND HIT ENTER TO CONNECT";
  } else if (isInputDrive) {
    placeholder = "Enter drive path (e.g. /Volumes/USB)...";
    headerLabel = "TYPE DRIVE PATH AND HIT ENTER TO CONNECT";
  } else if (isDriveExercises) {
    placeholder = "Search exercises on drive...";
    headerLabel = "SELECT EXERCISE TO FETCH FROM DRIVE";
  } else if (isRemoteExercises) {
    placeholder = "Search exercises on remote...";
    headerLabel = "SELECT EXERCISE TO FETCH FROM REMOTE";
  } else if (isInputExerciseID) {
    placeholder = "Enter Exercise ID to fetch (e.g. go101-lab01)...";
    headerLabel = "TYPE EXERCISE ID AND HIT ENTER";
  } else if (isInputExerciseVersion) {
    placeholder = "Enter version to fetch (optional, hit Enter for Latest)...";
    headerLabel = "TYPE VERSION AND HIT ENTER";
  }

  return (
    <div className="quick-open-overlay-blur">
      <div className="quick-open-container" ref={quickOpenRef}>
        {/* Wizard Static Title Header - Placed above the input row */}
        {headerLabel && (
          <div className="quick-open-wizard-header" style={{ padding: '8px 14px', fontSize: '0.72rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid var(--border-color)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', backgroundColor: 'var(--bg-card)' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{ opacity: 0.8 }}>
                {isExercises || isDriveExercises || isRemoteExercises ? (
                  <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" />
                ) : (
                  <path d="M21 12a9 9 0 0 1-9 9m9-9a9 9 0 0 0-9-9m9 9H3m9 9a9 9 0 0 1-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 0 1 9-9" />
                )}
              </svg>
              <span>{headerLabel}</span>
            </div>
          </div>
        )}

        <div className="quick-open-input-row">
          <input 
            type="text" 
            value={quickOpenPath} 
            onChange={(e) => setQuickOpenPath(e.target.value)}
            onKeyDown={handleQuickOpenKeyDown}
            placeholder={placeholder}
            className="quick-open-input"
            disabled={loading}
            autoFocus
          />
        </div>

        {/* Validation Errors inside Palette */}
        {validationError && (
          <div className="validation-alert-error" style={{ margin: '10px 18px 0 18px', textAlign: 'left' }}>
            {validationError}
          </div>
        )}

        {/* Parsed Workspace Manifest Preview - placed ABOVE folder names (only in browse mode) */}
        {isBrowse && quickOpenActiveManifest && (() => {
          const getLastDirName = (path) => {
            if (!path) return '';
            const cleanPath = path.replace(/[/\\]$/, '');
            return cleanPath.split(/[/\\]/).pop() || '';
          };
          const isDrive = !quickOpenActiveManifest.lab_id;
          const displayName = quickOpenActiveManifest.lab_id || getLastDirName(quickOpenPath);
          const displayTitle = quickOpenActiveManifest.title || (isDrive ? (quickOpenActiveManifest.owner ? `Drive Owner: ${quickOpenActiveManifest.owner}` : 'Drive Location') : '');
          
          return (
            <div 
              className={`quick-open-manifest-preview ${selectedIndex === manifestIdx ? 'active' : ''}`}
              onClick={handleConfirmQuickOpen} 
              onMouseEnter={() => setSelectedIndex(manifestIdx)}
              style={{ cursor: 'pointer' }}
              title={isDrive ? "Click to select this drive folder" : "Click to open this workspace"}
            >
              <div className="manifest-preview-icon">
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" className="tdes-logo-svg">
                  <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" />
                </svg>
              </div>
              <div className="manifest-preview-text">
                <div className="manifest-preview-id">
                  {displayName}
                  {quickOpenActiveManifest.version && (
                    <span className="manifest-preview-version">v{quickOpenActiveManifest.version}</span>
                  )}
                </div>
                <div className="manifest-preview-title">
                  {displayTitle}
                </div>
              </div>
            </div>
          );
        })()}

        <div className="quick-open-results-list">
          {isBrowse ? (
            <>
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
            </>
          ) : (
            selectableItems.map((item, idx) => {
              if (item.type === 'action') {
                return (
                  <div 
                    key={idx}
                    className={`quick-open-item ${selectedIndex === idx ? 'active' : ''}`}
                    onClick={() => {
                      setSelectedIndex(idx);
                      handleConfirmQuickOpen();
                    }}
                    onMouseEnter={() => setSelectedIndex(idx)}
                    style={{ borderBottom: '1px solid var(--border-color)', color: 'var(--primary)' }}
                  >
                    <span className="item-text font-semibold">{item.label}</span>
                  </div>
                );
              }
              if (item.type === 'input_confirm') {
                return (
                  <div 
                    key={idx}
                    className={`quick-open-item active`}
                    onClick={handleConfirmQuickOpen}
                    style={{ color: 'var(--primary)' }}
                  >
                    <span className="item-text font-mono font-semibold">{item.label}</span>
                  </div>
                );
              }
              if (item.type === 'exercise') {
                return (
                  <div 
                    key={idx}
                    className={`quick-open-item ${selectedIndex === idx ? 'active' : ''}`}
                    onClick={() => {
                      setSelectedIndex(idx);
                      handleConfirmQuickOpen();
                    }}
                    onMouseEnter={() => setSelectedIndex(idx)}
                  >
                    <div className="recent-details">
                      <span className="recent-name" style={{ fontSize: '0.85rem', display: 'flex', alignItems: 'center', gap: '8px' }}>
                        {item.data.lab_id}
                        {item.data.language && (
                          <span className="recent-lab-id" style={{ color: 'var(--primary)', backgroundColor: 'rgba(79, 70, 229, 0.05)', border: '1px solid rgba(79, 70, 229, 0.12)', textTransform: 'uppercase' }}>
                            {item.data.language}
                          </span>
                        )}
                        {item.data.latest && <span className="recent-lab-id" style={{ color: 'var(--accent)', backgroundColor: 'rgba(5, 150, 105, 0.05)', border: '1px solid rgba(5, 150, 105, 0.12)' }}>LATEST</span>}
                      </span>
                      <span className="recent-path" style={{ fontSize: '0.72rem' }}>
                        Version: {item.data.version} {item.data.title && `| ${item.data.title}`}
                      </span>
                    </div>
                  </div>
                );
              }
              // Render recent source item
              return (
                <div 
                  key={idx}
                  className={`quick-open-item ${selectedIndex === idx ? 'active' : ''}`}
                  onClick={() => {
                    setSelectedIndex(idx);
                    handleConfirmQuickOpen();
                  }}
                  onMouseEnter={() => setSelectedIndex(idx)}
                >
                  <div className="recent-details">
                    <span className="recent-name" style={{ fontSize: '0.85rem' }}>
                      {item.value}
                    </span>
                    <span className="recent-path" style={{ fontSize: '0.72rem', opacity: 0.6 }}>
                      Type: {item.sourceType === 'remote' ? 'Registry URL' : 'Drive Location'}
                    </span>
                  </div>
                </div>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
}
