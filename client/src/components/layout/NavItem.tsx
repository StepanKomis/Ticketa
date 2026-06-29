import { NavLink } from 'react-router-dom'

interface Props {
  icon: React.ReactNode
  label: string
  to?: string
  end?: boolean
  onClick?: () => void
  badge?: number
  disabled?: boolean
}

export default function NavItem({ icon, label, to, end, onClick, badge, disabled }: Props) {
  if (disabled) {
    return (
      <span className="sidebar__item sidebar__item--inert" aria-disabled="true">
        {icon}
        <span className="sidebar__label">{label}</span>
      </span>
    )
  }

  if (to) {
    return (
      <NavLink
        to={to}
        end={end}
        className={({ isActive }) => `sidebar__item${isActive ? ' sidebar__item--active' : ''}`}
      >
        {icon}
        <span className="sidebar__label">{label}</span>
        {badge != null && badge > 0 && <span className="sidebar__count">{badge}</span>}
      </NavLink>
    )
  }

  return (
    <button type="button" className="sidebar__logout" onClick={onClick}>
      {icon}
      <span className="sidebar__label">{label}</span>
    </button>
  )
}
