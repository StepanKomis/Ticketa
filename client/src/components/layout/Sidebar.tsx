import { Link } from 'react-router-dom'
import { LayoutDashboard, Ticket, Activity, BarChart2, Users, Settings, LogOut } from 'lucide-react'
import type { UserRole } from '../../types/ticket'
import { initials } from '../../utils/avatar'
import { ROLE_LABELS } from '../../utils/labels'
import NavItem from './NavItem'
import './Sidebar.scss'

const BrandGlyph = () => (
  <svg width="15" height="15" viewBox="0 0 100 100" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <line x1="50" y1="14" x2="50" y2="86" stroke="currentColor" strokeWidth="6" strokeDasharray="5 10" strokeLinecap="round" opacity="0.45"/>
    <path d="M30 33 H70 V44 H56 V70 H44 V44 H30 Z" fill="currentColor"/>
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
        <NavItem to="/" end icon={<LayoutDashboard size={18} strokeWidth={1.4} />} label="Přehled" />
        <NavItem to="/tickets" icon={<Ticket size={18} strokeWidth={1.4} />} label="Tikety" badge={ticketCount} />
        <NavItem to="/activity" icon={<Activity size={18} strokeWidth={1.4} />} label="Aktivita" />
        {isStaff && <NavItem disabled icon={<BarChart2 size={18} strokeWidth={1.4} />} label="Reporty" />}
        {role === 'staff' && <NavItem disabled icon={<Users size={18} strokeWidth={1.4} />} label="Adresář" />}
      </nav>

      <div className="sidebar__footer">
        {role === 'admin' && (
          <NavItem to="/settings" icon={<Settings size={18} strokeWidth={1.4} />} label="Nastavení" />
        )}
        {onLogout && (
          <NavItem onClick={onLogout} icon={<LogOut size={18} strokeWidth={1.4} />} label="Odhlásit se" />
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
