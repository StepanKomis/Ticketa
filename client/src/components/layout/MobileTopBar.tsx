import { Link } from 'react-router-dom'
import type { UserRole } from '../../types/ticket'
import { initials } from '../../utils/avatar'
import './MobileTopBar.css'

const BrandGlyph = () => (
  <svg width="13" height="13" viewBox="0 0 100 100" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <line x1="50" y1="14" x2="50" y2="86" stroke="currentColor" strokeWidth="6" strokeDasharray="5 10" strokeLinecap="round" opacity="0.45"/>
    <path d="M30 33 H70 V44 H56 V70 H44 V44 H30 Z" fill="currentColor"/>
  </svg>
)

const BellIcon = () => (
  <svg width="14.5" height="14.5" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <path d="M4 6.5a4 4 0 0 1 8 0c0 3 .8 4 1.4 4.5H2.6C3.2 10.5 4 9.5 4 6.5Z" stroke="currentColor" strokeWidth="1.3" strokeLinejoin="round"/>
    <path d="M6.5 13a1.5 1.5 0 0 0 3 0" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round"/>
  </svg>
)

interface Props {
  firstName?: string
  lastName?: string
  email: string
  role: UserRole
  unreadCount?: number
  onBellClick?: () => void
}

export default function MobileTopBar({ firstName, lastName, email, unreadCount = 0, onBellClick }: Props) {
  return (
    <header className="mobileTopBar">
      <Link to="/" className="mobileTopBar__brand">
        <span className="mobileTopBar__logo" aria-hidden="true"><BrandGlyph /></span>
        <span className="mobileTopBar__wordmark">Ticketa</span>
      </Link>
      <div className="mobileTopBar__actions">
        <button
          type="button"
          className="mobileTopBar__iconBtn"
          aria-label={unreadCount > 0 ? `Oznámení (${unreadCount} nových)` : 'Oznámení'}
          onClick={onBellClick}
        >
          <BellIcon />
          {unreadCount > 0 && (
            <span className="mobileTopBar__badge" aria-hidden="true">
              {unreadCount > 99 ? '99+' : unreadCount}
            </span>
          )}
        </button>
        <span className="mobileTopBar__avatar">{initials(firstName, lastName, email)}</span>
      </div>
    </header>
  )
}
