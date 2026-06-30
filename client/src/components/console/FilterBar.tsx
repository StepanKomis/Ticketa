import { useRef } from 'react'
import { Search } from 'lucide-react'
import type { TicketCategory } from '../../types/ticket'
import { PRIORITY_OPTIONS as PRIORITIES } from '../../utils/labels'
import './FilterBar.scss'

const CATEGORIES: TicketCategory[] = ['AV / Hardware', 'Síť / Internet', 'Nábytek', 'Budova / Prostory', 'Účty / Přístupy']

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
        <Search size={14} strokeWidth={1.4} />
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
