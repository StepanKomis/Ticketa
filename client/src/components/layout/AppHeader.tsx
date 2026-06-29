import type { ReactNode } from 'react'
import { Search, Bell, Plus } from 'lucide-react'
import type { UserRole } from '../../types/ticket'
import './AppHeader.scss'

interface Props {
  role: UserRole
  onNew?: () => void
  /** Replaces the default search field (e.g. a breadcrumb on the detail page). */
  left?: ReactNode
  showNew?: boolean
}

export default function AppHeader({ role, onNew, left, showNew = true }: Props) {
  const newLabel = role === 'student' ? 'Nový požadavek' : 'Nový tiket'

  return (
    <header className="appHeader">
      <div className="appHeader__left">
        {left ?? (
          <div className="appHeader__search" role="search">
            <Search size={14} strokeWidth={1.4} />
            <span className="appHeader__searchPlaceholder">Hledat tikety, místnosti, lidi…</span>
          </div>
        )}
      </div>

      <div className="appHeader__actions">
        <button type="button" className="appHeader__iconBtn" aria-label="Oznámení">
          <Bell size={15} strokeWidth={1.4} />
          <span className="appHeader__dot" aria-hidden="true" />
        </button>

        {showNew && (
          <button type="button" className="appHeader__newBtn" onClick={onNew}>
            <Plus size={13} strokeWidth={1.8} />
            {newLabel}
          </button>
        )}
      </div>
    </header>
  )
}
