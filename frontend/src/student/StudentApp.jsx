import React from 'react';
import { HashRouter, Routes, Route } from 'react-router-dom';
import { StudentProvider } from './StudentContext';
import Welcome from './pages/Welcome';
import ActiveWorkspace from './pages/ActiveWorkspace';
import CommandPalette from './components/CommandPalette';

export default function StudentApp() {
  return (
    <HashRouter>
      <StudentProvider>
        <div className="app-viewport">
          {/* Background Grid Lines (Teachyst Style) */}
          <div className="grid-lines-container">
            <div className="vertical-line v-line-1"></div>
            <div className="vertical-line v-line-2"></div>
            <div className="vertical-line v-line-3"></div>
          </div>

          {/* Main Content Area */}
          <main className="main-viewport-content">
            <Routes>
              <Route path="/" element={<Welcome />} />
              <Route path="/workspace" element={<ActiveWorkspace />} />
            </Routes>
          </main>

          {/* Command Palette Overlay */}
          <CommandPalette />
        </div>
      </StudentProvider>
    </HashRouter>
  );
}
