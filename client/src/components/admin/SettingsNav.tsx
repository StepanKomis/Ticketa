import { NavLink } from 'react-router-dom'
import { User, Users, Bell, Server } from 'lucide-react'
import { useAuth } from '../../hooks/useAuth'
import { usePendingCount } from '../../hooks/useUsers'
import './SettingsNav.scss'

const navClass = ({ isActive }: { isActive: boolean }) =>
  `settingsNav__item${isActive ? ' settingsNav__item--active' : ''}`

export default function SettingsNav() {
  const { user } = useAuth()
  const isAdmin = user?.role === 'admin'
  const pendingCount = usePendingCount(isAdmin)

  return (
    <nav className="settingsNav" aria-label="Nastavení">
      <NavLink to="/settings" end className={navClass}>
        <User size={16} strokeWidth={1.4} />
        <span>Profil</span>
      </NavLink>
      {isAdmin && (
        <NavLink to="/settings/users" className={navClass}>
          <Users size={16} strokeWidth={1.4} />
          <span>Uživatelé</span>
          {pendingCount > 0 && (
            <span className="settingsNav__badge" aria-label={`${pendingCount} čekají na schválení`} />
          )}
        </NavLink>
      )}
      <NavLink to="/settings/notifications" className={navClass}>
        <Bell size={16} strokeWidth={1.4} />
        <span>Oznámení</span>
      </NavLink>
      {isAdmin && (
        <NavLink to="/settings/server" className={navClass}>
          <Server size={16} strokeWidth={1.4} />
          <span>Server</span>
        </NavLink>
      )}
    </nav>
  )
}
