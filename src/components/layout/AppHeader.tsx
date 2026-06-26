import type { ReactNode } from 'react'
import type { UserRole } from '../../types/ticket'
import './AppHeader.css'

const SearchIcon = () => (
  <svg width="14" height="14" viewBox="0 0 14 14" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <circle cx="6" cy="6" r="4.5" stroke="currentColor" strokeWidth="1.3"/>
    <path d="M9.5 9.5 12 12" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round"/>
  </svg>
)

const BellIcon = () => (
  <svg width="15" height="15" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <path d="M4 6.5a4 4 0 0 1 8 0c0 3 .8 4 1.4 4.5H2.6C3.2 10.5 4 9.5 4 6.5Z" stroke="currentColor" strokeWidth="1.3" strokeLinejoin="round"/>
    <path d="M6.5 13a1.5 1.5 0 0 0 3 0" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round"/>
  </svg>
)

const PlusIcon = () => (
  <svg width="13" height="13" viewBox="0 0 13 13" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <path d="M6.5 2v9M2 6.5h9" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round"/>
  </svg>
)

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
            <SearchIcon />
            <span className="appHeader__searchPlaceholder">Hledat tikety, místnosti, lidi…</span>
          </div>
        )}
      </div>

      <div className="appHeader__actions">
        <button type="button" className="appHeader__iconBtn" aria-label="Oznámení">
          <BellIcon />
          <span className="appHeader__dot" aria-hidden="true" />
        </button>

        {showNew && (
          <button type="button" className="appHeader__newBtn" onClick={onNew}>
            <PlusIcon />
            {newLabel}
          </button>
        )}
      </div>
    </header>
  )
}
