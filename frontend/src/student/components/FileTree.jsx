import React, { useState } from 'react';

const FolderIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{ color: 'var(--primary)', opacity: 0.85 }}>
    <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
  </svg>
);

const FileIcon = () => (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" style={{ color: 'var(--text-secondary)', opacity: 0.75 }}>
    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
    <polyline points="14 2 14 8 20 8" />
  </svg>
);

const ChevronDown = () => (
  <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round" style={{ color: 'var(--text-muted)' }}>
    <polyline points="6 9 12 15 18 9" />
  </svg>
);

const ChevronRight = () => (
  <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round" style={{ color: 'var(--text-muted)' }}>
    <polyline points="9 18 15 12 9 6" />
  </svg>
);

function FileTreeNode({ node, onFileClick, activeFilePath }) {
  const [isExpanded, setIsExpanded] = useState(false);

  const handleClick = (e) => {
    e.stopPropagation();
    if (node.isDir) {
      setIsExpanded(!isExpanded);
    } else {
      onFileClick(node);
    }
  };

  const isActive = activeFilePath === node.path;

  return (
    <div className="file-tree-node" style={{ userSelect: 'none', textAlign: 'left' }}>
      <div 
        className={`file-tree-row ${isActive ? 'active' : ''}`}
        onClick={handleClick}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '6px',
          padding: '4px 8px',
          cursor: 'pointer',
          borderRadius: '4px',
          fontSize: '0.8rem',
          color: isActive ? 'var(--primary)' : 'var(--text-secondary)',
          backgroundColor: isActive ? 'rgba(79, 70, 229, 0.08)' : 'transparent',
          transition: 'all 0.15s ease'
        }}
      >
        {node.isDir ? (
          <div style={{ display: 'flex', alignItems: 'center', gap: '4px', width: '100%' }}>
            <span style={{ display: 'inline-flex', width: '12px', justifyContent: 'center' }}>
              {isExpanded ? <ChevronDown /> : <ChevronRight />}
            </span>
            <FolderIcon />
            <span style={{ fontWeight: 500 }}>{node.name}</span>
          </div>
        ) : (
          <div style={{ display: 'flex', alignItems: 'center', gap: '4px', paddingLeft: '16px' }}>
            <FileIcon />
            <span>{node.name}</span>
          </div>
        )}
      </div>

      {node.isDir && isExpanded && node.children && (
        <div className="file-tree-children" style={{ paddingLeft: '12px', borderLeft: '1px solid var(--border-color)', marginLeft: '13px', marginTop: '2px', marginBottom: '2px' }}>
          {node.children.map((child, idx) => (
            <FileTreeNode 
              key={idx} 
              node={child} 
              onFileClick={onFileClick} 
              activeFilePath={activeFilePath} 
            />
          ))}
        </div>
      )}
    </div>
  );
}

export default function FileTree({ files, onFileClick, activeFilePath }) {
  if (!files || files.length === 0) {
    return (
      <div style={{ padding: '12px', fontSize: '0.75rem', color: 'var(--text-muted)', fontStyle: 'italic', textAlign: 'center' }}>
        No files found
      </div>
    );
  }

  return (
    <div className="file-tree-container" style={{ display: 'flex', flexDirection: 'column', gap: '2px' }}>
      {files.map((node, idx) => (
        <FileTreeNode 
          key={idx} 
          node={node} 
          onFileClick={onFileClick} 
          activeFilePath={activeFilePath} 
        />
      ))}
    </div>
  );
}
