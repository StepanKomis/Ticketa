import { useState } from 'react'
import ConsoleLayout from '../components/layout/ConsoleLayout'
import HeroSection from '../components/console/HeroSection'
import StatCards from '../components/console/StatCards'
import ReportCTA from '../components/console/ReportCTA'
import TicketList from '../components/console/TicketList'
import ActivityFeed from '../components/console/ActivityFeed'
import Fab from '../components/layout/Fab'
import NewTicketModal from '../components/tickets/NewTicketModal'
import ResolveTicketModal from '../components/tickets/ResolveTicketModal'
import { useAuth } from '../hooks/useAuth'
import { useTickets, useTicketStatusCounts } from '../hooks/useTickets'
import { useStatuses } from '../hooks/useStatuses'
import { useTicketActions } from '../hooks/useTicketActions'
import type { ApiTicket } from '../types/api'
import type { Ticket } from '../types/ticket'
import type { FilterValue } from '../components/console/FilterTabs'
import { mapApiTicket, statusIdForUiStatus } from '../utils/mappers'
import { deriveActivity } from '../utils/activity'
import { timeGreeting, roleSubtitle } from '../utils/greetings'
import './consolePage.scss'

export default function ConsolePage() {
  const { user } = useAuth()
  const role = user?.role ?? 'student'
  const isStaff = role === 'staff' || role === 'admin'
  const isMaintainer = role === 'maintainer'

  const [filterTab, setFilterTab] = useState<FilterValue>('all')
  const [modalOpen, setModalOpen] = useState(false)

  const { data: statuses } = useStatuses()
  const { advance, resolveModal } = useTicketActions(statuses ?? [])

  // Stat card counts — independent of active tab
  const { data: allTickets } = useTickets({ limit: 1 })
  const statCounts = useTicketStatusCounts(statuses ?? [], allTickets?.total ?? 0)

  // Map active tab to server-side query params
  const tabParams = (() => {
    if (filterTab === 'resolved') return { closed: true as const }
    if (filterTab === 'in_progress') {
      const sid = statusIdForUiStatus('in_progress', statuses ?? [])
      return sid != null ? { status_id: sid } : { closed: false as const }
    }
    if (filterTab === 'open') {
      const sid = statusIdForUiStatus('open', statuses ?? [])
      return sid != null ? { status_id: sid } : { closed: false as const }
    }
    return {} // 'all' — no filter
  })()

  const { data: apiTickets, isLoading } = useTickets(tabParams)

  const tickets: Ticket[] = (apiTickets?.items ?? []).map((t: ApiTicket) => mapApiTicket(t, statuses ?? []))

  const tabCounts: Partial<Record<FilterValue, number>> = {
    all: allTickets?.total ?? 0,
    open: statCounts.open,
    in_progress: statCounts.inProgress,
    resolved: statCounts.resolved,
  }

  const firstName = user?.firstName || user?.email.split('@')[0] || ''
  const greeting = `${timeGreeting()}, ${firstName}`

  const handleNew = () => setModalOpen(true)
  const handleTicketAction = (isStaff || isMaintainer) ? advance : undefined
  const canAct = isMaintainer && !isStaff ? (t: Ticket) => t.assigneeId === user?.id : undefined

  return (
    <ConsoleLayout
      user={{ firstName: user?.firstName, lastName: user?.lastName, email: user?.email ?? '', role }}
      onNew={handleNew}
      fab={<Fab onClick={handleNew} label="Nový požadavek" />}
    >
      <div className="consolePage">
        <HeroSection greeting={greeting} subtitle={roleSubtitle(role)} className="consolePage__hero" />

        <StatCards
          open={statCounts.open}
          inProgress={statCounts.inProgress}
          resolved={statCounts.resolved}
        />

        <div className="consolePage__grid">
          <TicketList
            tickets={tickets}
            isLoading={isLoading}
            onTicketAction={handleTicketAction}
            canAct={canAct}
            filter={filterTab}
            onFilterChange={setFilterTab}
            counts={tabCounts}
          />

          <aside className="consolePage__side">
            <ReportCTA role={role} onNew={handleNew} />
            <ActivityFeed events={deriveActivity(tickets)} />
          </aside>
        </div>
      </div>

      <NewTicketModal open={modalOpen} role={role} onClose={() => setModalOpen(false)} />
      <ResolveTicketModal {...resolveModal} />
    </ConsoleLayout>
  )
}
