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
import { mapApiTicket } from '../utils/mappers'
import { deriveActivity } from '../utils/activity'
import { timeGreeting, roleSubtitle } from '../utils/greetings'
import './consolePage.css'

export default function ConsolePage() {
  const { user } = useAuth()
  const role = user?.role ?? 'student'
  const isStaff = role === 'staff' || role === 'admin'
  const isMaintainer = role === 'maintainer'

  // Dashboard zobrazuje jen aktivní (neuzavřené) tikety, ale StatCards
  // matematika (open = grandTotal - inProgress - resolved) potřebuje skutečný
  // celkový počet napříč všemi tikety — proto samostatný lehký dotaz.
  const { data: apiTickets, isLoading } = useTickets({ closed: false })
  const { data: allTickets } = useTickets({ limit: 1 })
  const { data: statuses } = useStatuses()
  const { advance, resolveModal } = useTicketActions(statuses ?? [])
  const statCounts = useTicketStatusCounts(statuses ?? [], allTickets?.total ?? 0)

  const [modalOpen, setModalOpen] = useState(false)

  const firstName = user?.firstName || user?.email.split('@')[0] || ''
  const greeting = `${timeGreeting()}, ${firstName}`

  const tickets: Ticket[] = (apiTickets?.items ?? []).map((t: ApiTicket) => mapApiTicket(t, statuses ?? []))

  const handleNew = () => setModalOpen(true)
  // Staff/admin can advance status on any ticket; maintainers only on
  // tickets assigned to them (enforced both here and server-side). Students
  // see no action buttons (no handler passed).
  const handleTicketAction = (isStaff || isMaintainer) ? advance : undefined
  const canAct = isMaintainer && !isStaff ? (t: Ticket) => t.assigneeId === user?.id : undefined

  return (
    <ConsoleLayout
      user={{ firstName: user?.firstName, lastName: user?.lastName, email: user?.email ?? '', role }}
      ticketCount={apiTickets?.total ?? tickets.length}
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
          <TicketList tickets={tickets} isLoading={isLoading} onTicketAction={handleTicketAction} canAct={canAct} />

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
