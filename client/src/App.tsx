import './App.scss';
import { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useNavigate, useLocation } from 'react-router-dom';
import { QueryClientProvider } from '@tanstack/react-query';
import queryClient from './lib/queryClient';
import { AuthProvider } from './contexts/AuthContext';
import { useAuth } from './contexts/AuthContext';
import ProtectedRoute from './components/auth/ProtectedRoute';
import AuthPage from './pages/authPage';
import ConsolePage from './pages/consolePage';
import TicketsPage from './pages/ticketsPage';
import ActivityPage from './pages/activityPage';
import TicketDetailPage from './pages/ticketDetailPage';
import SettingsPage from './pages/settingsPage';
import UsersPage from './pages/usersPage';
import NotificationsSettingsPage from './pages/notificationsSettingsPage';
import ServerSettingsPage from './pages/serverSettingsPage';
import AcceptInvitePage from './pages/AcceptInvitePage';
import WizardPage from './pages/wizardPage';
import { getSetupStatus } from './api/auth';

// Checks setup status once and redirects to /setup if wizard is not complete.
// Runs inside the router so navigate() is available.
function SetupGuard() {
  const navigate = useNavigate()
  const location = useLocation()
  const { user, isLoading } = useAuth()

  useEffect(() => {
    if (location.pathname === '/setup' || location.pathname.startsWith('/invite')) return
    getSetupStatus().then(status => {
      if (status.needs_setup || !status.wizard_completed) {
        // If admin exists but wizard not done, only redirect admin users (or unauthenticated first-run)
        if (status.needs_setup || (user?.role === 'admin' && !status.wizard_completed)) {
          navigate('/setup', { replace: true })
        }
      }
    }).catch(() => {})
  }, [navigate, location.pathname, user, isLoading])

  return null
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <BrowserRouter>
          <SetupGuard />
          <Routes>
            <Route path="/setup"    element={<WizardPage />} />
            <Route path="/"         element={<ProtectedRoute><ConsolePage /></ProtectedRoute>} />
            <Route path="/tickets"  element={<ProtectedRoute><TicketsPage /></ProtectedRoute>} />
            <Route path="/tickets/:id" element={<ProtectedRoute><TicketDetailPage /></ProtectedRoute>} />
            <Route path="/activity" element={<ProtectedRoute><ActivityPage /></ProtectedRoute>} />
            <Route path="/profile"        element={<Navigate to="/settings" replace />} />
            <Route path="/settings"       element={<ProtectedRoute><SettingsPage /></ProtectedRoute>} />
            <Route path="/settings/users" element={<ProtectedRoute roles={['admin']}><UsersPage /></ProtectedRoute>} />
            <Route path="/settings/notifications" element={<ProtectedRoute><NotificationsSettingsPage /></ProtectedRoute>} />
            <Route path="/settings/server" element={<ProtectedRoute roles={['admin']}><ServerSettingsPage /></ProtectedRoute>} />
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
