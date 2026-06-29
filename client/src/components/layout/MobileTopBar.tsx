import { Link } from 'react-router-dom'
import { Bell } from 'lucide-react'
import type { UserRole } from '../../types/ticket'
import { initials } from '../../utils/avatar'
import './MobileTopBar.scss'

const BrandGlyph = () => (
  <svg width="13" height="13" viewBox="0 0 100 100" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <line x1="50" y1="14" x2="50" y2="86" stroke="currentColor" strokeWidth="6" strokeDasharray="5 10" strokeLinecap="round" opacity="0.45"/>
    <path d="M30 33 H70 V44 H56 V70 H44 V44 H30 Z" fill="currentColor"/>
  </svg>
)

interface Props {
  firstName?: string
  lastName?: string
  email: string
  role: UserRole
}

export default function MobileTopBar({ firstName, lastName, email }: Props) {
  return (
    <header className="mobileTopBar">
      <Link to="/" className="mobileTopBar__brand">
        <span className="mobileTopBar__logo" aria-hidden="true"><BrandGlyph /></span>
        <span className="mobileTopBar__wordmark">Ticketa</span>
      </Link>
      <div className="mobileTopBar__actions">
        <button type="button" className="mobileTopBar__iconBtn" aria-label="Oznámení">
          <Bell size={15} strokeWidth={1.4} />
          <span className="mobileTopBar__dot" aria-hidden="true" />
        </button>
        <span className="mobileTopBar__avatar">{initials(firstName, lastName, email)}</span>
      </div>
    </header>
  )
}
