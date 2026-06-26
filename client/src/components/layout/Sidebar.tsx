import { Link, NavLink } from 'react-router-dom'
import type { UserRole } from '../../types/ticket'
import { initials } from '../../utils/avatar'
import './Sidebar.css'

const BrandGlyph = () => (
  <svg width="15" height="15" viewBox="0 0 100 100" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <line x1="50" y1="14" x2="50" y2="86" stroke="currentColor" strokeWidth="6" strokeDasharray="5 10" strokeLinecap="round" opacity="0.45"/>
    <path d="M30 33 H70 V44 H56 V70 H44 V44 H30 Z" fill="currentColor"/>
  </svg>
)

const OverviewIcon = () => (
  <svg width="18" height="15" viewBox="0 0 18 16" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <rect x="2" y="2" width="6" height="5" rx="1.2" stroke="currentColor" strokeWidth="1.3"/>
    <rect x="2" y="9" width="6" height="5" rx="1.2" stroke="currentColor" strokeWidth="1.3"/>
    <rect x="10" y="2" width="6" height="12" rx="1.2" stroke="currentColor" strokeWidth="1.3"/>
  </svg>
)

const TicketsIcon = () => (
  <svg width="18" height="15" viewBox="0 0 18 16" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <path d="M2.5 5A1.5 1.5 0 0 1 4 3.5h10A1.5 1.5 0 0 1 15.5 5v1.2a1.3 1.3 0 0 0 0 2.6V11A1.5 1.5 0 0 1 14 12.5H4A1.5 1.5 0 0 1 2.5 11V8.8a1.3 1.3 0 0 0 0-2.6V5Z" stroke="currentColor" strokeWidth="1.3" strokeLinejoin="round"/>
  </svg>
)

const ActivityIcon = () => (
  <svg width="18" height="15" viewBox="0 0 18 16" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <path d="M2 8h3l2-4 2.5 8L12 6l1.5 2H16" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

const ReportsIcon = () => (
  <svg width="18" height="15" viewBox="0 0 18 16" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <rect x="3" y="2" width="12" height="12" rx="2" stroke="currentColor" strokeWidth="1.3"/>
    <path d="M6 11V8M9 11V6M12 11v-2" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round"/>
  </svg>
)

const DirectoryIcon = () => (
  <svg width="18" height="15" viewBox="0 0 18 16" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <circle cx="7" cy="6" r="2.5" stroke="currentColor" strokeWidth="1.3"/>
    <path d="M2.5 13c0-2.2 2-3.6 4.5-3.6s4.5 1.4 4.5 3.6" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round"/>
    <path d="M12.5 4.2a2.3 2.3 0 0 1 0 4.3M13 13c0-1.6-.6-2.8-1.6-3.4" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round"/>
  </svg>
)

const SettingsIcon = () => (
  <svg width="18" height="15" viewBox="0 0 18 16" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <circle cx="9" cy="8" r="2.2" stroke="currentColor" strokeWidth="1.3"/>
    <path d="M9 1.8v1.6M9 12.6v1.6M3.4 4.6l1.1 1.1M13.5 10.3l1.1 1.1M1.8 8h1.6M14.6 8h1.6M3.4 11.4l1.1-1.1M13.5 5.7l1.1-1.1" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round"/>
  </svg>
)

const ROLE_LABELS: Record<UserRole, string> = {
  student: 'Student',
  staff: 'Učitel',
  maintainer: 'Údržbář',
  admin: 'Správce systému',
  pending: 'Čekající na schválení',
}

const LogoutIcon = () => (
  <svg width="18" height="15" viewBox="0 0 18 16" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <path d="M7 3H4a1 1 0 0 0-1 1v8a1 1 0 0 0 1 1h3" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round"/>
    <path d="M12 5.5 14.5 8 12 10.5M7.5 8h7" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" strokeLinejoin="round"/>
  </svg>
)

interface Props {
  firstName?: string
  lastName?: string
  email: string
  role: UserRole
  ticketCount?: number
  onLogout?: () => void
}

const navItemClass = ({ isActive }: { isActive: boolean }) =>
  `sidebar__item${isActive ? ' sidebar__item--active' : ''}`

export default function Sidebar({ firstName, lastName, email, role, ticketCount, onLogout }: Props) {
  const fullName = [firstName, lastName].filter(Boolean).join(' ') || email
  const isStaff = role === 'staff' || role === 'admin'

  return (
    <aside className="sidebar">
      <Link to="/" className="sidebar__brand">
        <span className="sidebar__logo" aria-hidden="true"><BrandGlyph /></span>
        <span className="sidebar__wordmark">Ticketa</span>
      </Link>

      <nav className="sidebar__nav" aria-label="Hlavní navigace">
        <NavLink to="/" end className={navItemClass}>
          <OverviewIcon />
          <span className="sidebar__label">Přehled</span>
        </NavLink>

        <NavLink to="/tickets" className={navItemClass}>
          <TicketsIcon />
          <span className="sidebar__label">Tikety</span>
          {ticketCount != null && ticketCount > 0 && (
            <span className="sidebar__count">{ticketCount}</span>
          )}
        </NavLink>

        <NavLink to="/activity" className={navItemClass}>
          <ActivityIcon />
          <span className="sidebar__label">Aktivita</span>
        </NavLink>

        {isStaff && (
          <span className="sidebar__item sidebar__item--inert" aria-disabled="true">
            <ReportsIcon />
            <span className="sidebar__label">Reporty</span>
          </span>
        )}

        {role === 'staff' && (
          <span className="sidebar__item sidebar__item--inert" aria-disabled="true">
            <DirectoryIcon />
            <span className="sidebar__label">Adresář</span>
          </span>
        )}
      </nav>

      <div className="sidebar__footer">
        {role === 'admin' && (
          <NavLink to="/settings" className={navItemClass}>
            <SettingsIcon />
            <span className="sidebar__label">Nastavení</span>
          </NavLink>
        )}

        {onLogout && (
          <button type="button" className="sidebar__item sidebar__logout" onClick={onLogout}>
            <LogoutIcon />
            <span className="sidebar__label">Odhlásit se</span>
          </button>
        )}

        <div className="sidebar__viewer">
          <span className="sidebar__viewerEyebrow">Zobrazeno jako</span>
          <div className="sidebar__viewerCard">
            <span className="sidebar__avatar">{initials(firstName, lastName, email)}</span>
            <div className="sidebar__viewerMeta">
              <span className="sidebar__viewerName">{fullName}</span>
              <span className="sidebar__viewerRole">{ROLE_LABELS[role]}</span>
            </div>
          </div>
        </div>
      </div>
    </aside>
  )
}
