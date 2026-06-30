import { useState, type ReactNode } from 'react'
import Sidebar from './Sidebar'
import AppHeader from './AppHeader'
import MobileTopBar from './MobileTopBar'
import BottomNav from './BottomNav'
import NotificationPanel from './NotificationPanel'
import { useAuth } from '../../hooks/useAuth'
import { useTickets } from '../../hooks/useTickets'
import { useNotifications, useMarkNotificationsViewed } from '../../hooks/useNotifications'
import type { UserRole } from '../../types/ticket'
import './ConsoleLayout.scss'

interface LayoutUser {
  firstName?: string
  lastName?: string
  email: string
  role: UserRole
}

interface Props {
  user: LayoutUser
  onNew?: () => void
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
  headerLeft,
  showNew = true,
  fab,
  children,
}: Props) {
  const { logout } = useAuth()
  const { data: openTickets } = useTickets({ closed: false, limit: 1 })
  const { data: notifData } = useNotifications()
  const markViewed = useMarkNotificationsViewed()
  const [panelOpen, setPanelOpen] = useState(false)

  const unreadCount = notifData?.unread_count ?? 0
  const notifications = notifData?.items ?? []

  const handleBellClick = () => {
    const opening = !panelOpen
    setPanelOpen(opening)
    if (opening && unreadCount > 0) {
      markViewed.mutate()
    }
  }

  return (
    <div className="appShell">
      <Sidebar
        firstName={user.firstName}
        lastName={user.lastName}
        email={user.email}
        role={user.role}
        ticketCount={openTickets?.total ?? 0}
        onLogout={logout}
      />

      <div className="appShell__body">
        <AppHeader
          role={user.role}
          onNew={onNew}
          left={headerLeft}
          showNew={showNew}
          unreadCount={unreadCount}
          onBellClick={handleBellClick}
        />
        <MobileTopBar
          firstName={user.firstName}
          lastName={user.lastName}
          email={user.email}
          role={user.role}
          unreadCount={unreadCount}
          onBellClick={handleBellClick}
        />

        <main className="appShell__main">{children}</main>
      </div>

      {panelOpen && (
        <NotificationPanel
          items={notifications}
          onClose={() => setPanelOpen(false)}
        />
      )}

      {fab && <div className="appShell__fab">{fab}</div>}
      <BottomNav />
    </div>
  )
}
