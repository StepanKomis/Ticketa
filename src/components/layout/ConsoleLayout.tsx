import type { ReactNode } from 'react'
import Sidebar from './Sidebar'
import AppHeader from './AppHeader'
import MobileTopBar from './MobileTopBar'
import BottomNav from './BottomNav'
import { useAuth } from '../../hooks/useAuth'
import type { UserRole } from '../../types/ticket'
import './ConsoleLayout.css'

interface LayoutUser {
  firstName?: string
  lastName?: string
  email: string
  role: UserRole
}

interface Props {
  user: LayoutUser
  onNew?: () => void
  ticketCount?: number
  /** Replaces the header search field (e.g. a breadcrumb on the detail page). */
  headerLeft?: ReactNode
  showNew?: boolean
  /** Floating action button rendered bottom-right on mobile. */
  fab?: ReactNode
  children: ReactNode
}

export default function ConsoleLayout({
  user,
  onNew,
  ticketCount,
  headerLeft,
  showNew = true,
  fab,
  children,
}: Props) {
  const { logout } = useAuth()

  return (
    <div className="appShell">
      <Sidebar
        firstName={user.firstName}
        lastName={user.lastName}
        email={user.email}
        role={user.role}
        ticketCount={ticketCount}
        onLogout={logout}
      />

      <div className="appShell__body">
        <AppHeader role={user.role} onNew={onNew} left={headerLeft} showNew={showNew} />
        <MobileTopBar
          firstName={user.firstName}
          lastName={user.lastName}
          email={user.email}
          role={user.role}
        />

        <main className="appShell__main">{children}</main>
      </div>

      {fab && <div className="appShell__fab">{fab}</div>}
      <BottomNav />
    </div>
  )
}
