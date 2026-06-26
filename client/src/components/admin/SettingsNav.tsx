import { NavLink } from 'react-router-dom'
import { useAuth } from '../../hooks/useAuth'
import { usePendingCount } from '../../hooks/useUsers'
import './SettingsNav.css'

const ProfileIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
    <circle cx="8" cy="6" r="2.6" stroke="currentColor" strokeWidth="1.3" />
    <path d="M3 13c0-2.4 2.2-3.9 5-3.9s5 1.5 5 3.9" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" />
  </svg>
)

const UsersIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
    <circle cx="6" cy="6" r="2.2" stroke="currentColor" strokeWidth="1.3" />
    <path d="M2 13c0-2 1.8-3.3 4-3.3S10 11 10 13" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" />
    <path d="M11 4.2a2 2 0 0 1 0 3.8M14 13c0-1.4-.6-2.5-1.6-3.1" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" />
  </svg>
)

const BellIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
    <path d="M4 6.5a4 4 0 0 1 8 0c0 3 .8 4 1.4 4.5H2.6C3.2 10.5 4 9.5 4 6.5Z" stroke="currentColor" strokeWidth="1.3" strokeLinejoin="round" />
    <path d="M6.5 13a1.5 1.5 0 0 0 3 0" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" />
  </svg>
)

const LockIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
    <rect x="3" y="7" width="10" height="7" rx="1.5" stroke="currentColor" strokeWidth="1.3" />
    <path d="M5 7V5a3 3 0 0 1 6 0v2" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" />
    <circle cx="8" cy="10.5" r="1" fill="currentColor" />
  </svg>
)

const MailIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
    <rect x="2" y="4" width="12" height="9" rx="1.5" stroke="currentColor" strokeWidth="1.3" />
    <path d="M2.5 4.8 8 9l5.5-4.2" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

const navClass = ({ isActive }: { isActive: boolean }) =>
  `settingsNav__item${isActive ? ' settingsNav__item--active' : ''}`

export default function SettingsNav() {
  const { user } = useAuth()
  const pendingCount = usePendingCount(user?.role === 'admin')

  return (
    <nav className="settingsNav" aria-label="Nastavení">
      <NavLink to="/settings" end className={navClass}>
        <ProfileIcon />
        <span>Profil</span>
      </NavLink>
      <NavLink to="/settings/users" className={navClass}>
        <UsersIcon />
        <span>Uživatelé</span>
        {pendingCount > 0 && (
          <span className="settingsNav__badge" aria-label={`${pendingCount} čekají na schválení`} />
        )}
      </NavLink>
      <NavLink to="/settings/password" className={navClass}>
        <LockIcon />
        <span>Heslo</span>
      </NavLink>
      <NavLink to="/settings/email" className={navClass}>
        <MailIcon />
        <span>E-mail</span>
      </NavLink>
      <span className="settingsNav__item settingsNav__item--inert" aria-disabled="true">
        <BellIcon />
        <span>Oznámení</span>
      </span>
    </nav>
  )
}
