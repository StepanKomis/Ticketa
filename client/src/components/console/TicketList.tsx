import FilterTabs from './FilterTabs'
import type { FilterValue } from './FilterTabs'
import TicketCard from './TicketCard'
import { useTicketFilter, matchesFilter } from '../../hooks/useTicketFilter'
import type { Ticket } from '../../types/ticket'
import './TicketList.css'

interface Props {
  tickets: Ticket[]
  title?: string
  isLoading?: boolean
  onTicketAction?: (ticket: Ticket) => void
  canAct?: (ticket: Ticket) => boolean
  // Controlled mode — server already filtered the tickets; skip client-side re-filtering
  filter?: FilterValue
  onFilterChange?: (v: FilterValue) => void
  counts?: Partial<Record<FilterValue, number>>
}

const FILTERS: FilterValue[] = ['all', 'open', 'in_progress', 'resolved']

export default function TicketList({ tickets, title = 'Moje tikety', isLoading, onTicketAction, canAct, filter: filterProp, onFilterChange, counts: countsProp }: Props) {
  const internal = useTicketFilter(tickets)

  const controlled = filterProp !== undefined
  const filter = controlled ? filterProp : internal.filter
  const setFilter = controlled ? (onFilterChange ?? (() => {})) : internal.setFilter
  const filtered = controlled ? tickets : internal.filtered

  const counts = countsProp ?? Object.fromEntries(
    FILTERS.map(f => [f, tickets.filter(t => matchesFilter(t, f)).length]),
  ) as Record<FilterValue, number>

  return (
    <section className="ticketList">
      <div className="ticketList__header">
        <h2 className="ticketList__title">{title}</h2>
        <FilterTabs active={filter} onChange={setFilter} counts={counts} />
      </div>

      {isLoading ? (
        <ul className="ticketList__rows" aria-busy="true">
          {[1, 2, 3].map(n => (
            <li key={n} className="ticketList__skeleton" aria-hidden="true" />
          ))}
        </ul>
      ) : filtered.length === 0 ? (
        <p className="ticketList__empty">Žádné tikety v tomto filtru.</p>
      ) : (
        <ul className="ticketList__rows">
          {filtered.map(ticket => (
            <li key={ticket.id} className="ticketList__row">
              <TicketCard ticket={ticket} onAction={onTicketAction} canAct={canAct} />
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
