import { NavLink } from 'react-router-dom'
import { Home, FileText, Activity, User } from 'lucide-react'
import './BottomNav.scss'

export default function BottomNav() {
  const navClass = ({ isActive }: { isActive: boolean }) =>
    `bottomNav__item${isActive ? ' bottomNav__item--active' : ''}`

  return (
    <nav className="bottomNav" aria-label="Mobilní navigace">
      <NavLink to="/"         className={navClass} end>
        <Home size={20} strokeWidth={1.4} />
        <span>Přehled</span>
      </NavLink>
      <NavLink to="/tickets"  className={navClass}>
        <FileText size={20} strokeWidth={1.4} />
        <span>Tikety</span>
      </NavLink>
      <NavLink to="/activity" className={navClass}>
        <Activity size={20} strokeWidth={1.4} />
        <span>Aktivita</span>
      </NavLink>
      <NavLink to="/settings" className={navClass}>
        <User size={20} strokeWidth={1.4} />
        <span>Profil</span>
      </NavLink>
    </nav>
  )
}
