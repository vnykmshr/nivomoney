/**
 * Admin App
 * Main application component with routing
 * Security-focused: All routes require admin authentication
 */

import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AdminRoute } from './components/AdminRoute';
import { AdminLogin } from './pages/AdminLogin';
import { AdminDashboard } from './pages/AdminDashboard';
import { AdminKYC } from './pages/AdminKYC';
import { Users } from './pages/Users';
import { UserDetail } from './pages/UserDetail';
import { Transactions } from './pages/Transactions';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Public route - Login only */}
        <Route path="/login" element={<AdminLogin />} />

        {/* Protected admin routes */}
        <Route
          path="/"
          element={
            <AdminRoute>
              <AdminDashboard />
            </AdminRoute>
          }
        />

        <Route
          path="/kyc"
          element={
            <AdminRoute>
              <AdminKYC />
            </AdminRoute>
          }
        />

        <Route
          path="/users"
          element={
            <AdminRoute>
              <Users />
            </AdminRoute>
          }
        />

        <Route
          path="/users/:userId"
          element={
            <AdminRoute>
              <UserDetail />
            </AdminRoute>
          }
        />

        <Route
          path="/transactions"
          element={
            <AdminRoute>
              <Transactions />
            </AdminRoute>
          }
        />

        {/* Future routes (Phase 5):
          - /reports (compliance reports)
          - /settings (admin settings)
          - /audit (audit log viewer)
        */}

        {/* Catch-all redirect */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
