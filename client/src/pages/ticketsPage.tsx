import { useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import ConsoleLayout from '../components/layout/ConsoleLayout'
import TicketList from '../components/console/TicketList'
import FilterBar, { FilterBarValues } from '../components/console/FilterBar'
import type { FilterValue } from '../components/console/FilterTabs'
import NewTicketModal from '../components/tickets/NewTicketModal'
import ResolveTicketModal from '../components/tickets/ResolveTicketModal'
import Fab from '../components/layout/Fab'
import { useAuth } from '../hooks/useAuth'
import { useTickets } from '../hooks/useTickets'
import { useStatuses } from '../hooks/useStatuses'
import { useTicketActions } from '../hooks/useTicketActions'
import { mapApiTicket } from '../utils/mappers'
import { statusIdForUiStatus } from '../utils/mappers'
import type { ApiTicket } from '../types/api'
import type { Ticket } from '../types/ticket'
import './ticketsPage.scss'

const PAGE_SIZE = 20

type MaintainerScope = 'mine' | 'unassigned' | 'all'

export default function TicketsPage() {
  const { user } = useAuth()
  const role = user?.role ?? 'student'
  const isStaff = role === 'staff' || role === 'admin'
  const isMaintainer = role === 'maintainer'

  const [searchParams, setSearchParams] = useSearchParams()
  const [modalOpen, setModalOpen] = useState(false)

  const statusFilter = (searchParams.get('status') ?? 'all') as FilterValue
  const filterQ        = searchParams.get('q') ?? ''
  const filterPriority = searchParams.get('priority') ?? ''
  const filterCategory = searchParams.get('category') ?? ''
  const filterOffset   = Number(searchParams.get('offset') ?? '0')
  const scope = (searchParams.get('scope') ?? 'mine') as MaintainerScope
  const showDeleted = isStaff && searchParams.get('show_deleted') === 'true'

  const { data: statuses } = useStatuses()
  const { advance, resolveModal } = useTicketActions(statuses ?? [])

  // "Vše" = aktivní tikety (uzavřené se nezobrazují, dokud o ně uživatel
  // explicitně nepožádá přes tab "Vyřešené", který teď cílí přímo na
  // is_closed namísto odvozování přes status_id).
  const statusIdParam = (statusFilter !== 'all' && statusFilter !== 'resolved')
    ? statusIdForUiStatus(statusFilter, statuses ?? [])
    : undefined
  const closedParam = statusFilter === 'all' ? false : statusFilter === 'resolved' ? true : undefined

  const { data: ticketList, isLoading } = useTickets({
    ...(showDeleted
      ? { show_deleted: true }
      : {
          status_id: statusIdParam,
          closed:    closedParam,
          ...(isMaintainer && scope === 'mine' ? { assigned_to: user?.id } : {}),
          ...(isMaintainer && scope === 'unassigned' ? { unassigned: true } : {}),
        }
    ),
    priority:  filterPriority || undefined,
    category:  filterCategory || undefined,
    q:         filterQ || undefined,
    limit:     PAGE_SIZE,
    offset:    filterOffset,
  })

  const apiTickets = ticketList?.items ?? []
  const total = ticketList?.total ?? 0
  const tickets = apiTickets.map((t: ApiTicket) => mapApiTicket(t, statuses ?? []))

  function setParam(key: string, value: string) {
    setSearchParams(prev => {
      const next = new URLSearchParams(prev)
      if (value) {
        next.set(key, value)
      } else {
        next.delete(key)
      }
      next.delete('offset')
      return next
    })
  }

  function handleStatusChange(val: FilterValue) {
    setParam('status', val === 'all' ? '' : val)
  }

  function handleScopeChange(val: MaintainerScope) {
    setParam('scope', val === 'mine' ? '' : val)
  }

  function handleFilterChange(vals: FilterBarValues) {
    setSearchParams(prev => {
      const next = new URLSearchParams(prev)
      vals.q         ? next.set('q', vals.q)               : next.delete('q')
      vals.priority  ? next.set('priority', vals.priority)  : next.delete('priority')
      vals.category  ? next.set('category', vals.category)  : next.delete('category')
      next.delete('offset')
      return next
    })
  }

  function goToPage(offset: number) {
    setSearchParams(prev => {
      const next = new URLSearchParams(prev)
      offset > 0 ? next.set('offset', String(offset)) : next.delete('offset')
      return next
    })
  }

  const handleNew = () => setModalOpen(true)
  const handleTicketAction = (isStaff || isMaintainer) ? advance : undefined
  const canAct = isMaintainer && !isStaff ? (t: Ticket) => t.assigneeId === user?.id : undefined
  const totalPages = Math.ceil(total / PAGE_SIZE)
  const currentPage = Math.floor(filterOffset / PAGE_SIZE) + 1
  const listTitle = isMaintainer
    ? (scope === 'mine' ? 'Moje přiřazené' : scope === 'unassigned' ? 'Nepřiřazené' : 'Všechny tikety')
    : (role === 'student' ? 'Moje tikety' : 'Všechny tikety')

  return (
    <ConsoleLayout
      user={{ firstName: user?.firstName, lastName: user?.lastName, email: user?.email ?? '', role }}
      onNew={handleNew}
      fab={<Fab onClick={handleNew} label={role === 'student' ? 'Nový požadavek' : 'Nový tiket'} />}
    >
      <div className="ticketsPage__content">
        <div className="ticketsPage__header">
          <h1 className="ticketsPage__title">Tikety</h1>
          {!isLoading && (
            <p className="ticketsPage__count">{total} tiketů</p>
          )}
        </div>

        {isMaintainer && (
          <div className="filterTabs" role="tablist" aria-label="Rozsah tiketů">
            {([
              { value: 'mine', label: 'Moje přiřazené' },
              { value: 'unassigned', label: 'Nepřiřazené' },
              { value: 'all', label: 'Všechny' },
            ] as { value: MaintainerScope; label: string }[]).map(tab => (
              <button
                key={tab.value}
                type="button"
                role="tab"
                className={`filterTabs__tab${scope === tab.value ? ' filterTabs__tab--active' : ''}`}
                onClick={() => handleScopeChange(tab.value)}
                aria-selected={scope === tab.value}
              >
                {tab.label}
              </button>
            ))}
          </div>
        )}

        {isStaff && (
          <div className="ticketsPage__deletedToggle">
            <button
              type="button"
              className={`ticketsPage__deletedBtn${showDeleted ? ' ticketsPage__deletedBtn--active' : ''}`}
              onClick={() => setParam('show_deleted', showDeleted ? '' : 'true')}
            >
              {showDeleted ? '← Aktivní tikety' : 'Smazané tikety'}
            </button>
          </div>
        )}

        {!showDeleted && (
          <FilterBar
            values={{ q: filterQ, priority: filterPriority, category: filterCategory }}
            onChange={handleFilterChange}
          />
        )}

        <TicketList
          tickets={tickets}
          title={showDeleted ? 'Smazané tikety' : listTitle}
          isLoading={isLoading}
          onTicketAction={handleTicketAction}
          canAct={canAct}
          filter={statusFilter}
          onFilterChange={handleStatusChange}
          counts={{ [statusFilter]: total }}
        />

        {totalPages > 1 && (
          <div className="ticketsPage__pagination">
            <button
              type="button"
              className="ticketsPage__pageBtn"
              disabled={currentPage <= 1}
              onClick={() => goToPage(filterOffset - PAGE_SIZE)}
            >
              ← Předchozí
            </button>
            <span className="ticketsPage__pageInfo">{currentPage} / {totalPages}</span>
            <button
              type="button"
              className="ticketsPage__pageBtn"
              disabled={currentPage >= totalPages}
              onClick={() => goToPage(filterOffset + PAGE_SIZE)}
            >
              Další →
            </button>
          </div>
        )}
      </div>

      <NewTicketModal open={modalOpen} role={role} onClose={() => setModalOpen(false)} />
      <ResolveTicketModal {...resolveModal} />
    </ConsoleLayout>
  )
}
