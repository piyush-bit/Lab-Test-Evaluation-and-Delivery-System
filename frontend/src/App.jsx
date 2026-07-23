import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { StudentProvider } from './student/StudentContext';
import { StudentAppContent } from './student/StudentApp';
import AdminPage from './admin/AdminPage';

export default function App() {
  return (
    <BrowserRouter>
      <StudentProvider>
        <Routes>
          {/* Admin Routes - Accessible directly via /admin, /admin/drive, etc. */}
          <Route path="/admin" element={<AdminPage />} />
          <Route path="/admin/drive" element={<AdminPage />} />
          <Route path="/admin/*" element={<AdminPage />} />

          {/* Student Routes */}
          <Route path="/*" element={<StudentAppContent />} />
        </Routes>
      </StudentProvider>
    </BrowserRouter>
  );
}
