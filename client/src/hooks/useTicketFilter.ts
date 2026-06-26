import { useMemo, useState } from 'react'
import type { FilterValue } from '../components/console/FilterTabs'
import type { Ticket } from '../types/ticket'

interface UseTicketFilterResult {
  filter: FilterValue
  setFilter: (value: FilterValue) => void
  filtered: Ticket[]
}

// Groups the five UI statuses into the three buckets the design exposes:
// "Otevřené" covers new + open, "Vyřešené" covers resolved + closed.
export function matchesFilter(ticket: Ticket, filter: FilterValue): boolean {
  switch (filter) {
    case 'all':
      return true
    case 'open':
      return ticket.status === 'new' || ticket.status === 'open'
    case 'in_progress':
      return ticket.status === 'in_progress'
    case 'resolved':
      return ticket.status === 'resolved' || ticket.status === 'closed'
    default:
      return true
  }
}

export function useTicketFilter(tickets: Ticket[]): UseTicketFilterResult {
  const [filter, setFilter] = useState<FilterValue>('all')

  const filtered = useMemo<Ticket[]>(
    () => tickets.filter(t => matchesFilter(t, filter)),
    [tickets, filter],
  )

  return { filter, setFilter, filtered }
}
