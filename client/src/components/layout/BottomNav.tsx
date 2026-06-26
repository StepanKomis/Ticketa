import { NavLink } from 'react-router-dom'
import './BottomNav.css'

const HomeIcon = () => (
  <svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M3 8.5L10 3l7 5.5V17a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V8.5z" stroke="currentColor" strokeWidth="1.4" fill="none"/>
    <path d="M7.5 18V13h5v5" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round"/>
  </svg>
)

const TicketsIcon = () => (
  <svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
    <rect x="3" y="4" width="14" height="12" rx="2" stroke="currentColor" strokeWidth="1.4" fill="none"/>
    <path d="M7 8h6M7 11.5h4" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round"/>
  </svg>
)

const ActivityIcon = () => (
  <svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
    <path d="M2 10h3l2.5-5.5L10 14.5l2.5-8L15 10h3" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round" fill="none"/>
  </svg>
)

const ProfileIcon = () => (
  <svg width="20" height="20" viewBox="0 0 20 20" fill="none" xmlns="http://www.w3.org/2000/svg">
    <circle cx="10" cy="7" r="3.5" stroke="currentColor" strokeWidth="1.4" fill="none"/>
    <path d="M3 18c0-3.314 3.134-6 7-6s7 2.686 7 6" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" fill="none"/>
  </svg>
)

export default function BottomNav() {
  const navClass = ({ isActive }: { isActive: boolean }) =>
    `bottomNav__item${isActive ? ' bottomNav__item--active' : ''}`

  return (
    <nav className="bottomNav" aria-label="Mobilní navigace">
      <NavLink to="/"         className={navClass} end>
        <HomeIcon />
        <span>Přehled</span>
      </NavLink>
      <NavLink to="/tickets"  className={navClass}>
        <TicketsIcon />
        <span>Tikety</span>
      </NavLink>
      <NavLink to="/activity" className={navClass}>
        <ActivityIcon />
        <span>Aktivita</span>
      </NavLink>
      <NavLink to="/profile"  className={navClass}>
        <ProfileIcon />
        <span>Profil</span>
      </NavLink>
    </nav>
  )
}
