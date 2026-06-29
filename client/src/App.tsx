import './App.scss';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import queryClient from './lib/queryClient';
import { AuthProvider } from './contexts/AuthContext';
import ProtectedRoute from './components/auth/ProtectedRoute';
import AuthPage from './pages/authPage';
import ConsolePage from './pages/consolePage';
import TicketsPage from './pages/ticketsPage';
import ActivityPage from './pages/activityPage';
import TicketDetailPage from './pages/ticketDetailPage';
import SettingsPage from './pages/settingsPage';
import UsersPage from './pages/usersPage';
import AcceptInvitePage from './pages/AcceptInvitePage';

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/"         element={<ProtectedRoute><ConsolePage /></ProtectedRoute>} />
            <Route path="/tickets"  element={<ProtectedRoute><TicketsPage /></ProtectedRoute>} />
            <Route path="/tickets/:id" element={<ProtectedRoute><TicketDetailPage /></ProtectedRoute>} />
            <Route path="/activity" element={<ProtectedRoute><ActivityPage /></ProtectedRoute>} />
            <Route path="/profile"        element={<Navigate to="/settings" replace />} />
            <Route path="/settings"       element={<ProtectedRoute><SettingsPage /></ProtectedRoute>} />
            <Route path="/settings/users" element={<ProtectedRoute roles={['admin']}><UsersPage /></ProtectedRoute>} />
            <Route path="/settings/password" element={<Navigate to="/settings" replace />} />
            <Route path="/settings/email" element={<Navigate to="/settings" replace />} />
            <Route path="/login"    element={<AuthPage form="login" />} />
            <Route path="/register" element={<AuthPage form="register" />} />
            <Route path="/invite/accept" element={<AcceptInvitePage />} />
            <Route path="*"         element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </QueryClientProvider>
  );
}

export default App;
