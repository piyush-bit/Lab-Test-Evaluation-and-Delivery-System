import React from 'react';
import { useStudent } from '../StudentContext';
import { DriveIcon, ExerciseIcon, OpenFolderIcon, WrenchIcon, HistoryIcon } from './Icons';

// Declarative Configuration Registry for Command Palette Modes
const PALETTE_CONFIGS = {
  admin_drive_menu: {
    headerLabel: 'ADMIN DRIVE MANAGEMENT',
    placeholder: 'Type action (Prepare a Drive, Open Drive) or select disk...',
    iconType: 'drive',
    isFolderBrowse: false,
  },
  open_workspace: {
    headerLabel: 'OPEN EXISTING EXERCISE WORKSPACE',
    placeholder: 'Navigate to an initialized exercise workspace folder...',
    iconType: 'exercise',
    isFolderBrowse: true,
  },
  admin_browse_prepare: {
    headerLabel: 'PREPARE A NEW TDES DRIVE',
    placeholder: 'Navigate or type target folder path to prepare...',
    iconType: 'wrench',
    isFolderBrowse: true,
  },
  admin_browse_open: {
    headerLabel: 'OPEN AN EXISTING TDES DRIVE',
    placeholder: 'Navigate to an existing TDES Drive folder...',
    iconType: 'drive',
    isFolderBrowse: true,
  },
  browse: {
    headerLabel: 'CHOOSE INITIALIZATION TARGET DIRECTORY',
    placeholder: 'Navigate or type target folder path to unpack exercise...',
    iconType: 'folder',
    isFolderBrowse: true,
  },
  exercises: {
    headerLabel: 'SELECT EXERCISE TO INITIALIZE',
    placeholder: 'Search cached exercises...',
    iconType: 'exercise',
    isFolderBrowse: false,
  },
  select_source: {
    headerLabel: 'SELECT REPOSITORY SOURCE OR DRIVE LOCATION',
    placeholder: 'Search or select repository/drive location...',
    iconType: 'folder',
    isFolderBrowse: false,
  },
  input_remote: {
    headerLabel: 'TYPE REGISTRY URL AND HIT ENTER TO CONNECT',
    placeholder: 'Enter remote registry URL (e.g. http://localhost:8080)...',
    iconType: 'folder',
    isFolderBrowse: false,
  },
  input_drive: {
    headerLabel: 'TYPE DRIVE PATH AND HIT ENTER TO CONNECT',
    placeholder: 'Enter drive path (e.g. /Volumes/USB)...',
    iconType: 'drive',
    isFolderBrowse: false,
  },
  drive_exercises: {
    headerLabel: 'SELECT EXERCISE TO FETCH FROM DRIVE',
    placeholder: 'Search exercises on drive...',
    iconType: 'exercise',
    isFolderBrowse: false,
  },
  remote_exercises: {
    headerLabel: 'SELECT EXERCISE TO FETCH FROM REMOTE',
    placeholder: 'Search exercises on remote...',
    iconType: 'exercise',
    isFolderBrowse: false,
  },
  input_exercise_id: {
    headerLabel: 'TYPE EXERCISE ID AND HIT ENTER',
    placeholder: 'Enter Exercise ID to fetch (e.g. go101-lab01)...',
    iconType: 'exercise',
    isFolderBrowse: false,
  },
  input_exercise_version: {
    headerLabel: 'TYPE VERSION AND HIT ENTER',
    placeholder: 'Enter version to fetch (optional, hit Enter for Latest)...',
    iconType: 'exercise',
    isFolderBrowse: false,
  },
  admin_add_exercise: {
    headerLabel: 'SELECT CACHED EXERCISE TO DEPLOY TO DRIVE',
    placeholder: 'Search cached exercises to deploy to drive...',
    iconType: 'exercise',
    isFolderBrowse: false,
  },
  admin_confirm_clear_results: {
    headerLabel: 'CLEAR ALL EVALUATION RESULTS?',
    placeholder: 'Select Yes to confirm clearing all evaluation results...',
    iconType: 'wrench',
    isFolderBrowse: false,
  },
  admin_confirm_clear_submissions: {
    headerLabel: 'PERMANENTLY CLEAR ALL SUBMISSIONS FROM DRIVE?',
    placeholder: 'Select Yes to confirm deleting all student submission files...',
    iconType: 'wrench',
    isFolderBrowse: false,
  },
};

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
    selectedIndex,
    setSelectedIndex,
    quickOpenParent,
    handleGoUp,
    filteredDirs,
    handleQuickOpenNavigate,
    getPathParts,
    quickOpenMode,
    selectableItems,
    validationError,
    loading,
    quickOpenSelectedExercise
  } = useStudent();

  if (!showQuickOpen) return null;

  // Resolve config dynamically for current mode
  const currentConfig = PALETTE_CONFIGS[quickOpenMode] || PALETTE_CONFIGS.browse;
  const isFolderBrowse = currentConfig.isFolderBrowse;

  const renderHeaderIcon = (iconType) => {
    switch (iconType) {
      case 'drive':
        return <DriveIcon size={14} />;
      case 'exercise':
        return <ExerciseIcon size={14} />;
      case 'wrench':
        return <WrenchIcon size={14} />;
      default:
        return <OpenFolderIcon size={14} />;
    }
  };

  return (
    <div className="quick-open-overlay-blur">
      <div className="quick-open-container" ref={quickOpenRef}>
        {/* Wizard Header Bar */}
        {currentConfig.headerLabel && (
          <div className="quick-open-wizard-header" style={{ padding: '8px 14px', fontSize: '0.72rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em', borderBottom: '1px solid var(--border-color)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', backgroundColor: 'var(--bg-card)' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
              {renderHeaderIcon(currentConfig.iconType)}
              <span>{currentConfig.headerLabel}</span>
            </div>
          </div>
        )}

        {/* Active Selected Exercise Bar (if step 2) */}
        {quickOpenSelectedExercise && (
          <div className="quick-open-selected-exercise-row" style={{ padding: '12px 16px', borderBottom: '1px solid var(--border-color)', backgroundColor: 'var(--bg-main)', textAlign: 'left' }}>
            <div style={{ fontSize: '0.68rem', fontWeight: 700, color: 'var(--text-muted)', marginBottom: '6px', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              Selected Exercise
            </div>
            <div className="recent-details">
              <span className="recent-name" style={{ fontSize: '0.85rem', display: 'flex', alignItems: 'center', gap: '8px', fontWeight: 600 }}>
                <ExerciseIcon size={14} /> {quickOpenSelectedExercise.lab_id}
                {quickOpenSelectedExercise.language && (
                  <span className="recent-lab-id" style={{ color: 'var(--primary)', backgroundColor: 'rgba(79, 70, 229, 0.05)', border: '1px solid rgba(79, 70, 229, 0.12)', textTransform: 'uppercase' }}>
                    {quickOpenSelectedExercise.language}
                  </span>
                )}
                {quickOpenSelectedExercise.latest && <span className="recent-lab-id" style={{ color: 'var(--accent)', backgroundColor: 'rgba(5, 150, 105, 0.05)', border: '1px solid rgba(5, 150, 105, 0.12)' }}>LATEST</span>}
              </span>
              <span className="recent-path" style={{ fontSize: '0.72rem', color: 'var(--text-secondary)', marginTop: '2px', display: 'block' }}>
                Version: {quickOpenSelectedExercise.version} {quickOpenSelectedExercise.title && `| ${quickOpenSelectedExercise.title}`}
              </span>
            </div>
          </div>
        )}

        {/* Input Row */}
        <div className="quick-open-input-row">
          <input 
            type="text" 
            value={quickOpenPath} 
            onChange={(e) => setQuickOpenPath(e.target.value)}
            onKeyDown={handleQuickOpenKeyDown}
            placeholder={currentConfig.placeholder}
            className="quick-open-input"
            disabled={loading}
            autoFocus
          />
        </div>

        {/* Validation Errors */}
        {validationError && (
          <div className="validation-alert-error" style={{ margin: '10px 18px 0 18px', textAlign: 'left' }}>
            {validationError}
          </div>
        )}

        {/* Manifest Preview Card - Atomic & Mode-Strict */}
        {isFolderBrowse && quickOpenActiveManifest && (() => {
          const isDrive = quickOpenActiveManifest.is_drive;
          const isWorkspace = quickOpenActiveManifest.is_workspace;
          const isBrowsePrepare = quickOpenMode === 'admin_browse_prepare';
          const isBrowseOpen = quickOpenMode === 'admin_browse_open';
          const isStudentBrowse = quickOpenMode === 'browse';

          // Strictly render manifest preview card ONLY for Category A modes (Entity Opening)
          if (quickOpenMode === 'admin_browse_open' && !isDrive) return null;
          if (quickOpenMode === 'open_workspace' && !isWorkspace) return null;
          if (quickOpenMode !== 'admin_browse_open' && quickOpenMode !== 'open_workspace') return null;

          const getLastDirName = (path) => {
            if (!path) return '';
            const cleanPath = path.replace(/[/\\]$/, '');
            return cleanPath.split(/[/\\]/).pop() || '';
          };

          const displayName = isWorkspace ? quickOpenActiveManifest.lab_id : getLastDirName(quickOpenPath);
          const displayTitle = isWorkspace 
            ? (quickOpenActiveManifest.title || 'Exercise Workspace')
            : (quickOpenActiveManifest.owner ? `Owner: ${quickOpenActiveManifest.owner}` : 'TDES Delivery Drive');
          
          return (
            <div 
              className={`quick-open-manifest-preview ${selectedIndex === manifestIdx ? 'active' : ''}`}
              onClick={handleConfirmQuickOpen} 
              onMouseEnter={() => {
                if (manifestIdx !== -1) setSelectedIndex(manifestIdx);
              }}
              style={{ cursor: 'pointer' }}
              title={isDrive ? "Click to open this drive" : "Click to open this workspace"}
            >
              <div className="manifest-preview-icon">
                {isDrive ? <DriveIcon size={22} /> : <ExerciseIcon size={22} />}
              </div>
              <div className="manifest-preview-text">
                <div className="manifest-preview-id" style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  {displayName}
                  {quickOpenActiveManifest.version && (
                    <span className="manifest-preview-version">v{quickOpenActiveManifest.version}</span>
                  )}
                  {isDrive && <span className="recent-lab-id" style={{ color: 'var(--accent)', backgroundColor: 'rgba(5, 150, 105, 0.05)', border: '1px solid rgba(5, 150, 105, 0.12)' }}>TDES DRIVE</span>}
                </div>
                <div className="manifest-preview-title">
                  {displayTitle}
                </div>
              </div>
            </div>
          );
        })()}

        {/* Results Items List */}
        <div className="quick-open-results-list">
          {selectableItems.map((item, idx) => {
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

            if (item.type === 'select_current_dir') {
              return (
                <div 
                  key={idx}
                  className={`quick-open-item ${selectedIndex === idx ? 'active' : ''}`}
                  onClick={() => {
                    setSelectedIndex(idx);
                    handleConfirmQuickOpen();
                  }}
                  onMouseEnter={() => setSelectedIndex(idx)}
                  style={{ borderBottom: '1px solid var(--border-color)', color: 'var(--primary)', fontWeight: 600 }}
                >
                  <div className="recent-details">
                    <span className="recent-name" style={{ fontSize: '0.85rem', display: 'flex', alignItems: 'center', gap: '6px' }}>
                      <OpenFolderIcon size={14} /> {item.label || `Select directory: ${item.path.split('/').pop() || item.path.split('\\').pop() || item.path}`}
                    </span>
                    <span className="recent-path" style={{ fontSize: '0.72rem', opacity: 0.6 }}>
                      {item.path}
                    </span>
                  </div>
                </div>
              );
            }

            if (item.type === 'input_confirm') {
              return (
                <div 
                  key={idx}
                  className="quick-open-item active"
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
                      <ExerciseIcon size={14} /> {item.data.lab_id}
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

            if (item.type === 'recent_drive_disk') {
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
                    <span className="recent-name" style={{ fontSize: '0.85rem', display: 'flex', alignItems: 'center', gap: '6px' }}>
                      <DriveIcon size={14} /> {item.value.split('/').pop() || item.value.split('\\').pop()}
                    </span>
                    <span className="recent-path" style={{ fontSize: '0.72rem', opacity: 0.6 }}>
                      {item.value}
                    </span>
                  </div>
                </div>
              );
            }

            if (item.type === 'up') {
              return (
                <div 
                  key={idx}
                  className={`quick-open-item go-up ${selectedIndex === idx ? 'active' : ''}`} 
                  onClick={handleGoUp}
                  onMouseEnter={() => setSelectedIndex(idx)}
                >
                  <span className="item-text">..</span>
                </div>
              );
            }

            if (item.type === 'dir') {
              return (
                <div 
                  key={idx}
                  className={`quick-open-item ${selectedIndex === idx ? 'active' : ''}`}
                  onClick={() => {
                    const { baseDir, sep } = getPathParts(quickOpenPath);
                    handleQuickOpenNavigate(baseDir + item.name + sep);
                  }}
                  onMouseEnter={() => setSelectedIndex(idx)}
                >
                  <span className="item-text">{item.name}</span>
                </div>
              );
            }

            if (item.type === 'recent_source') {
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
                    <span className="recent-name" style={{ fontSize: '0.85rem', display: 'flex', alignItems: 'center', gap: '6px' }}>
                      {item.sourceType === 'drive' ? <DriveIcon size={14} /> : <OpenFolderIcon size={14} />} {item.value}
                    </span>
                    <span className="recent-path" style={{ fontSize: '0.72rem', opacity: 0.6 }}>
                      Type: {item.sourceType === 'remote' ? 'Registry URL' : 'Drive Location'}
                    </span>
                  </div>
                </div>
              );
            }

            return null;
          })}
        </div>
      </div>
    </div>
  );
}
