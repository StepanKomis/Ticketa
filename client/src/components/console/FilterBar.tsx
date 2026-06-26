import { useRef } from 'react'
import type { TicketCategory, TicketPriority } from '../../types/ticket'
import './FilterBar.css'

const CATEGORIES: TicketCategory[] = ['AV / Hardware', 'Síť / Internet', 'Nábytek', 'Budova / Prostory', 'Účty / Přístupy']
const PRIORITIES: { value: TicketPriority; label: string }[] = [
  { value: 'low',    label: 'Nízká' },
  { value: 'medium', label: 'Střední' },
  { value: 'high',   label: 'Vysoká' },
  { value: 'urgent', label: 'Urgentní' },
]

const SearchIcon = () => (
  <svg width="14" height="14" viewBox="0 0 14 14" fill="none" aria-hidden="true">
    <circle cx="6" cy="6" r="4" stroke="currentColor" strokeWidth="1.3"/>
    <path d="m9 9 2.5 2.5" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round"/>
  </svg>
)

export interface FilterBarValues {
  q: string
  priority: string
  category: string
}

interface Props {
  values: FilterBarValues
  onChange: (next: FilterBarValues) => void
}

export default function FilterBar({ values, onChange }: Props) {
  const searchRef = useRef<HTMLInputElement>(null)
  const hasFilters = values.q || values.priority || values.category

  function set(key: keyof FilterBarValues, val: string) {
    onChange({ ...values, [key]: val })
  }

  function reset() {
    onChange({ q: '', priority: '', category: '' })
    searchRef.current?.focus()
  }

  return (
    <div className="filterBar">
      <label className="filterBar__search">
        <SearchIcon />
        <input
          ref={searchRef}
          className="filterBar__searchInput"
          type="search"
          placeholder="Hledat tikety…"
          value={values.q}
          onChange={e => set('q', e.target.value)}
          aria-label="Hledat tikety"
        />
      </label>

      <select
        className="filterBar__select"
        value={values.category}
        onChange={e => set('category', e.target.value)}
        aria-label="Kategorie"
      >
        <option value="">Kategorie</option>
        {CATEGORIES.map(c => (
          <option key={c} value={c}>{c}</option>
        ))}
      </select>

      <select
        className="filterBar__select"
        value={values.priority}
        onChange={e => set('priority', e.target.value)}
        aria-label="Priorita"
      >
        <option value="">Priorita</option>
        {PRIORITIES.map(p => (
          <option key={p.value} value={p.value}>{p.label}</option>
        ))}
      </select>

      {hasFilters && (
        <button type="button" className="filterBar__reset" onClick={reset} aria-label="Zrušit filtry">
          Zrušit
        </button>
      )}
    </div>
  )
}
