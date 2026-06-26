import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '../../hooks/useAuth'
import type { UserRole } from '../../types/ticket'

interface Props {
  children: React.ReactNode
  /** When set, only these roles may enter; others are redirected home. */
  roles?: UserRole[]
}

export default function ProtectedRoute({ children, roles }: Props) {
  const { user, isLoading } = useAuth()
  const { pathname } = useLocation()

  if (isLoading) return null

  if (!user) return <Navigate to="/login" replace />

  // Lokální uživatelé s nucenou změnou hesla jsou přesměrováni na /settings/password.
  if (user.mustChangePw && pathname !== '/settings/password') {
    return <Navigate to="/settings/password" replace />
  }

  if (roles && !roles.includes(user.role)) return <Navigate to="/" replace />

  return <>{children}</>
}
