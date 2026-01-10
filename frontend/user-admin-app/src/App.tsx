/**
 * User-Admin Portal App
 * Verification portal for User-Admin accounts
 */

import { Routes, Route, Navigate } from 'react-router-dom';
import { ProtectedRoute } from './components/ProtectedRoute';
import { Login } from './pages/Login';
import { Dashboard } from './pages/Dashboard';

function App() {
  return (
    <Routes>
      {/* Public route - Login */}
      <Route path="/login" element={<Login />} />

      {/* Protected routes */}
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <Dashboard />
          </ProtectedRoute>
        }
      />

      {/* Catch-all redirect */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

export default App;
