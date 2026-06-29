import { useState } from 'react'
import { usePatchTicket } from './useTickets'
import { statusIdForUiStatus } from '../utils/mappers'
import type { ApiTicketStatus } from '../types/api'
import type { Ticket, TicketStatus } from '../types/ticket'

// Maps a ticket's current status to the status its quick-action advances to:
// "Zahájit" (new → in_progress) and "Vyřešit" (in_progress → resolved).
function nextStatus(status: TicketStatus): TicketStatus | null {
  if (status === 'new' || status === 'open') return 'in_progress'
  if (status === 'in_progress') return 'resolved'
  if (status === 'resolved' || status === 'closed') return 'open'
  return null
}

// Quick status transitions used by ticket rows. Uses PATCH (canMetaUpdate /
// canUpdateOwnStatus on the backend) instead of PUT — PUT is gated by
// canEditContent (admin or unassigned author only), so staff/maintainer
// quick actions on tickets they didn't author would otherwise 403. Advancing
// to "resolved" requires a resolution note first — the caller must render
// <ResolveTicketModal {...resolveModal} />, since this hook has no JSX of
// its own and is shared by multiple pages.
export function useTicketActions(statuses: ApiTicketStatus[]) {
  const patch = usePatchTicket()
  const [pendingTicket, setPendingTicket] = useState<Ticket | null>(null)

  const advance = (ticket: Ticket) => {
    const target = nextStatus(ticket.status)
    if (!target) return
    const statusId = statusIdForUiStatus(target, statuses)
    if (statusId == null) return
    if (target === 'resolved') {
      setPendingTicket(ticket)
      return
    }
    patch.mutate({ id: Number(ticket.id), payload: { status_id: statusId } })
  }

  const confirmResolve = (note: string) => {
    if (!pendingTicket) return
    const statusId = statusIdForUiStatus('resolved', statuses)
    if (statusId == null) return
    patch.mutate(
      { id: Number(pendingTicket.id), payload: { status_id: statusId, resolution_note: note } },
      { onSuccess: () => setPendingTicket(null) },
    )
  }

  return {
    advance,
    isPending: patch.isPending,
    resolveModal: {
      open: pendingTicket !== null,
      onClose: () => setPendingTicket(null),
      onConfirm: confirmResolve,
      isPending: patch.isPending,
    },
  }
}
