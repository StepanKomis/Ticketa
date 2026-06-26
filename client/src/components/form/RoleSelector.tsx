import './RoleSelector.css'

type Role = 'student' | 'staff' | 'maintainer'

const ROLES: { value: Role; label: string }[] = [
  { value: 'student',    label: 'Student' },
  { value: 'staff',      label: 'Zaměstnanec' },
  { value: 'maintainer', label: 'Údržbář' },
]

interface Props {
  value: Role
  onChange: (role: Role) => void
  disabled?: boolean
  /** Pokud true, selector je uzamčen na roli admin (první uživatel v systému). */
  lockedToAdmin?: boolean
}

export default function RoleSelector({ value, onChange, disabled, lockedToAdmin }: Props) {
  return (
    <div className="roleSelector">
      <span className="roleSelector__label">Jsem…</span>
      <div className={`roleSelector__buttons${lockedToAdmin ? ' roleSelector__buttons--locked' : ''}`} role="group" aria-label="Vyberte roli">
        {ROLES.map(r => (
          <button
            key={r.value}
            type="button"
            className={`roleSelector__btn${value === r.value && !lockedToAdmin ? ' roleSelector__btn--selected' : ''}${lockedToAdmin ? ' roleSelector__btn--locked' : ''}`}
            onClick={() => !lockedToAdmin && onChange(r.value)}
            aria-pressed={value === r.value && !lockedToAdmin}
            disabled={disabled || lockedToAdmin}
            tabIndex={lockedToAdmin ? -1 : undefined}
          >
            {r.label}
          </button>
        ))}
        {lockedToAdmin && (
          <div className="roleSelector__adminOverlay" aria-live="polite">
            <span className="roleSelector__adminBadge">Admin</span>
            <span className="roleSelector__adminNote">Roli přidělí systém automaticky</span>
          </div>
        )}
      </div>
    </div>
  )
}
