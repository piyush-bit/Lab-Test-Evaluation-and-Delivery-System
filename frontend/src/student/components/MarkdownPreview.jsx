import React, { useState } from 'react';

// Simple, robust Markdown parser and renderer component
export default function MarkdownPreview({ content, isDarkMode }) {
  const [copiedCodeIdx, setCopiedCodeIdx] = useState(null);

  if (!content) {
    return (
      <div style={{ padding: '32px', textAlign: 'center', color: 'var(--text-muted)' }}>
        No content to preview
      </div>
    );
  }

  const handleCopyCode = (code, idx) => {
    navigator.clipboard.writeText(code);
    setCopiedCodeIdx(idx);
    setTimeout(() => setCopiedCodeIdx(null), 1500);
  };

  // Helper to render inline formatting: bold, italic, inline code, links
  const renderInline = (text) => {
    if (typeof text !== 'string') return text;

    // Process inline code blocks first to avoid regex conflicts
    const parts = [];
    const inlineCodeRegex = /`([^`]+)`/g;
    let lastIdx = 0;
    let match;

    while ((match = inlineCodeRegex.exec(text)) !== null) {
      if (match.index > lastIdx) {
        parts.push(text.slice(lastIdx, match.index));
      }
      parts.push(
        <code
          key={`code-${match.index}`}
          style={{
            fontFamily: 'var(--font-mono)',
            fontSize: '0.85em',
            padding: '2px 6px',
            borderRadius: '4px',
            backgroundColor: isDarkMode ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.06)',
            color: isDarkMode ? '#e879f9' : '#8b5cf6',
            border: '1px solid var(--border-color)'
          }}
        >
          {match[1]}
        </code>
      );
      lastIdx = match.index + match[0].length;
    }
    if (lastIdx < text.length) {
      parts.push(text.slice(lastIdx));
    }

    // Process bold, italic, and links inside non-code text parts
    return parts.map((part, pIdx) => {
      if (typeof part !== 'string') return part;

      // Simple link parser: [text](url)
      const linkRegex = /\[([^\]]+)\]\(([^)]+)\)/g;
      const subParts = [];
      let lLastIdx = 0;
      let lMatch;

      while ((lMatch = linkRegex.exec(part)) !== null) {
        if (lMatch.index > lLastIdx) {
          subParts.push(part.slice(lLastIdx, lMatch.index));
        }
        subParts.push(
          <a
            key={`link-${pIdx}-${lMatch.index}`}
            href={lMatch[2]}
            target="_blank"
            rel="noopener noreferrer"
            style={{ color: 'var(--primary)', textDecoration: 'underline' }}
          >
            {lMatch[1]}
          </a>
        );
        lLastIdx = lMatch.index + lMatch[0].length;
      }
      if (lLastIdx < part.length) {
        subParts.push(part.slice(lLastIdx));
      }

      return subParts.map((sub, sIdx) => {
        if (typeof sub !== 'string') return sub;

        // Bold formatting **text**
        const boldRegex = /\*\*([^*]+)\*\*/g;
        const bParts = [];
        let bLast = 0;
        let bMatch;

        while ((bMatch = boldRegex.exec(sub)) !== null) {
          if (bMatch.index > bLast) {
            bParts.push(sub.slice(bLast, bMatch.index));
          }
          bParts.push(<strong key={`b-${pIdx}-${sIdx}-${bMatch.index}`}>{bMatch[1]}</strong>);
          bLast = bMatch.index + bMatch[0].length;
        }
        if (bLast < sub.length) {
          bParts.push(sub.slice(bLast));
        }

        return bParts;
      });
    });
  };

  // Block-level parsing
  const lines = content.split(/\r?\n/);
  const elements = [];
  let inCodeBlock = false;
  let codeBlockLang = '';
  let codeBlockLines = [];
  let codeBlockCounter = 0;

  let inTable = false;
  let tableHeaders = [];
  let tableRows = [];

  const flushTable = (key) => {
    if (!inTable) return;
    elements.push(
      <div key={key} style={{ overflowX: 'auto', margin: '16px 0' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.9rem' }}>
          <thead>
            <tr style={{ backgroundColor: isDarkMode ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.03)', borderBottom: '2px solid var(--border-color)' }}>
              {tableHeaders.map((th, idx) => (
                <th key={idx} style={{ padding: '8px 12px', textAlign: 'left', fontWeight: 600 }}>
                  {renderInline(th.trim())}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {tableRows.map((row, rIdx) => (
              <tr key={rIdx} style={{ borderBottom: '1px solid var(--border-color)', backgroundColor: rIdx % 2 === 1 ? (isDarkMode ? 'rgba(255,255,255,0.02)' : 'rgba(0,0,0,0.01)') : 'transparent' }}>
                {row.map((td, cIdx) => (
                  <td key={cIdx} style={{ padding: '8px 12px' }}>
                    {renderInline(td.trim())}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
    inTable = false;
    tableHeaders = [];
    tableRows = [];
  };

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];

    // Code block toggle (```)
    if (line.trim().startsWith('```')) {
      if (inCodeBlock) {
        // Close code block
        const codeText = codeBlockLines.join('\n');
        const codeIdx = codeBlockCounter++;
        elements.push(
          <div
            key={`codeblock-${i}`}
            style={{
              margin: '16px 0',
              borderRadius: '8px',
              border: '1px solid var(--border-color)',
              backgroundColor: isDarkMode ? '#0a0a0c' : '#f8fafc',
              overflow: 'hidden'
            }}
          >
            <div style={{
              display: 'flex',
              justify: 'space-between',
              alignItems: 'center',
              padding: '6px 12px',
              backgroundColor: isDarkMode ? 'rgba(255,255,255,0.04)' : 'rgba(0,0,0,0.04)',
              borderBottom: '1px solid var(--border-color)',
              fontSize: '0.75rem',
              color: 'var(--text-muted)'
            }}>
              <span style={{ fontFamily: 'var(--font-mono)', fontWeight: 600, textTransform: 'uppercase' }}>
                {codeBlockLang || 'code'}
              </span>
              <button
                onClick={() => handleCopyCode(codeText, codeIdx)}
                style={{
                  background: 'none',
                  border: 'none',
                  cursor: 'pointer',
                  color: copiedCodeIdx === codeIdx ? 'var(--accent)' : 'var(--text-secondary)',
                  fontSize: '0.75rem',
                  fontWeight: 500,
                  display: 'flex',
                  alignItems: 'center',
                  gap: '4px'
                }}
              >
                {copiedCodeIdx === codeIdx ? '✓ Copied' : 'Copy'}
              </button>
            </div>
            <pre className="monaco-style-scrollbar" style={{
              margin: 0,
              padding: '12px 16px',
              overflowX: 'auto',
              fontFamily: 'var(--font-mono)',
              fontSize: '0.85rem',
              lineHeight: 1.6,
              color: isDarkMode ? '#e2e8f0' : '#1e293b'
            }}>
              <code>{codeText}</code>
            </pre>
          </div>
        );
        inCodeBlock = false;
        codeBlockLines = [];
        codeBlockLang = '';
      } else {
        flushTable(`tbl-${i}`);
        inCodeBlock = true;
        codeBlockLang = line.trim().slice(3).trim();
        codeBlockLines = [];
      }
      continue;
    }

    if (inCodeBlock) {
      codeBlockLines.push(line);
      continue;
    }

    // Table rows (| col | col |)
    if (line.trim().startsWith('|') && line.trim().endsWith('|')) {
      const cells = line.trim().slice(1, -1).split('|');
      if (!inTable) {
        inTable = true;
        tableHeaders = cells;
      } else if (cells.every(c => c.trim().match(/^:?-+:?$/))) {
        // Table separator row line, skip
      } else {
        tableRows.push(cells);
      }
      continue;
    } else {
      flushTable(`tbl-${i}`);
    }

    // Headings
    if (line.startsWith('# ')) {
      elements.push(
        <h1 key={`h1-${i}`} style={{ fontSize: '1.75rem', fontWeight: 700, margin: '24px 0 12px 0', borderBottom: '1px solid var(--border-color)', paddingBottom: '8px', color: 'var(--text-primary)' }}>
          {renderInline(line.slice(2))}
        </h1>
      );
      continue;
    }
    if (line.startsWith('## ')) {
      elements.push(
        <h2 key={`h2-${i}`} style={{ fontSize: '1.35rem', fontWeight: 600, margin: '20px 0 10px 0', borderBottom: '1px solid var(--border-color)', paddingBottom: '6px', color: 'var(--text-primary)' }}>
          {renderInline(line.slice(3))}
        </h2>
      );
      continue;
    }
    if (line.startsWith('### ')) {
      elements.push(
        <h3 key={`h3-${i}`} style={{ fontSize: '1.1rem', fontWeight: 600, margin: '16px 0 8px 0', color: 'var(--text-primary)' }}>
          {renderInline(line.slice(4))}
        </h3>
      );
      continue;
    }
    if (line.startsWith('#### ')) {
      elements.push(
        <h4 key={`h4-${i}`} style={{ fontSize: '0.95rem', fontWeight: 600, margin: '14px 0 6px 0', color: 'var(--text-primary)' }}>
          {renderInline(line.slice(5))}
        </h4>
      );
      continue;
    }

    // Horizontal Rule
    if (line.trim() === '---' || line.trim() === '***' || line.trim() === '___') {
      elements.push(<hr key={`hr-${i}`} style={{ border: 'none', borderTop: '1px solid var(--border-color)', margin: '20px 0' }} />);
      continue;
    }

    // Blockquote
    if (line.startsWith('> ')) {
      elements.push(
        <blockquote key={`bq-${i}`} style={{
          margin: '12px 0',
          padding: '8px 16px',
          borderLeft: '4px solid var(--primary)',
          backgroundColor: isDarkMode ? 'rgba(99, 102, 241, 0.08)' : 'rgba(99, 102, 241, 0.05)',
          color: 'var(--text-secondary)',
          borderRadius: '0 4px 4px 0'
        }}>
          {renderInline(line.slice(2))}
        </blockquote>
      );
      continue;
    }

    // Unordered List Items
    if (line.trim().startsWith('- ') || line.trim().startsWith('* ')) {
      const listContent = line.trim().slice(2);
      // Check for Task List items [- ] or [x]
      if (listContent.startsWith('[ ] ') || listContent.startsWith('[x] ') || listContent.startsWith('[X] ')) {
        const checked = listContent.startsWith('[x] ') || listContent.startsWith('[X] ');
        elements.push(
          <div key={`task-${i}`} style={{ display: 'flex', alignItems: 'center', gap: '8px', margin: '4px 0', fontSize: '0.9rem', color: 'var(--text-primary)' }}>
            <input type="checkbox" checked={checked} readOnly style={{ accentColor: 'var(--primary)' }} />
            <span style={{ textDecoration: checked ? 'line-through' : 'none', color: checked ? 'var(--text-muted)' : 'inherit' }}>
              {renderInline(listContent.slice(4))}
            </span>
          </div>
        );
      } else {
        elements.push(
          <li key={`li-${i}`} style={{ margin: '4px 0 4px 20px', fontSize: '0.92rem', color: 'var(--text-primary)', lineHeight: 1.6 }}>
            {renderInline(listContent)}
          </li>
        );
      }
      continue;
    }

    // Ordered List Items
    const olMatch = line.trim().match(/^(\d+)\.\s+(.*)$/);
    if (olMatch) {
      elements.push(
        <div key={`ol-${i}`} style={{ display: 'flex', gap: '8px', margin: '4px 0 4px 16px', fontSize: '0.92rem', color: 'var(--text-primary)', lineHeight: 1.6 }}>
          <span style={{ fontWeight: 600, color: 'var(--primary)', fontFamily: 'var(--font-mono)' }}>{olMatch[1]}.</span>
          <span>{renderInline(olMatch[2])}</span>
        </div>
      );
      continue;
    }

    // Empty lines
    if (line.trim() === '') {
      elements.push(<div key={`br-${i}`} style={{ height: '8px' }} />);
      continue;
    }

    // Regular paragraphs
    elements.push(
      <p key={`p-${i}`} style={{ margin: '6px 0', lineHeight: 1.65, fontSize: '0.93rem', color: 'var(--text-primary)' }}>
        {renderInline(line)}
      </p>
    );
  }

  return (
    <div className="markdown-preview-container monaco-style-scrollbar" style={{
      padding: '24px 32px',
      width: '100%',
      maxWidth: '100%',
      margin: 0,
      boxSizing: 'border-box',
      overflowY: 'auto',
      height: '100%',
      color: 'var(--text-primary)'
    }}>
      {elements}
    </div>
  );
}
