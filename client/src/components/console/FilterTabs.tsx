import './FilterTabs.scss'

export type FilterValue = 'all' | 'open' | 'in_progress' | 'resolved' | 'deleted'

interface Tab {
  value: FilterValue
  label: string
}

const TABS: Tab[] = [
  { value: 'all',         label: 'Vše' },
  { value: 'open',        label: 'Otevřené' },
  { value: 'in_progress', label: 'Řeší se' },
  { value: 'resolved',    label: 'Vyřešené' },
]

const DELETED_TAB: Tab = { value: 'deleted', label: 'Smazané' }

interface Props {
  active: FilterValue
  onChange: (value: FilterValue) => void
  counts?: Partial<Record<FilterValue, number>>
  showDeletedTab?: boolean
}

export default function FilterTabs({ active, onChange, counts, showDeletedTab }: Props) {
  const tabs = showDeletedTab ? [...TABS, DELETED_TAB] : TABS
  return (
    <div className="filterTabs" role="tablist" aria-label="Filtrovat tikety">
      {tabs.map(tab => {
        const count = counts?.[tab.value]
        const isActive = active === tab.value
        return (
          <button
            key={tab.value}
            type="button"
            role="tab"
            className={`filterTabs__tab${isActive ? ' filterTabs__tab--active' : ''}${tab.value === 'deleted' ? ' filterTabs__tab--deleted' : ''}`}
            onClick={() => onChange(tab.value)}
            aria-selected={isActive}
          >
            {tab.label}
            {count != null && <span className="filterTabs__count">{count}</span>}
          </button>
        )
      })}
    </div>
  )
}
