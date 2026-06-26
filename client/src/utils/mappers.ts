import type { ApiTicket, ApiTicketStatus } from '../types/api'
import type { Ticket, TicketStatus } from '../types/ticket'

// Formats a ticket id for display as "TK-1234", tolerating ids that are
// already prefixed.
export function formatTicketId(id: string): string {
  return /^tk-/i.test(id) ? id : `TK-${id}`
}

export function mapApiTicket(
  apiTicket: ApiTicket,
  statuses: ApiTicketStatus[] = [],
): Ticket {
  return {
    id: String(apiTicket.ID),
    title: apiTicket.Title,
    body: apiTicket.Body,
    status: resolveStatus(apiTicket.StatusID, statuses),
    createdAt: new Date(apiTicket.CreatedAt),
    updatedAt: apiTicket.UpdatedAt ? new Date(apiTicket.UpdatedAt) : undefined,
    authorId: apiTicket.AuthorID,
    authorName: apiTicket.AuthorName || undefined,
    priority: (apiTicket.Priority as Ticket['priority']) || undefined,
    requestedPriority: (apiTicket.RequestedPriority as Ticket['requestedPriority']) || null,
    category: (apiTicket.Category as Ticket['category']) || undefined,
    location: apiTicket.Location || undefined,
    assigneeId: apiTicket.AssignedTo ?? null,
    assigneeName: apiTicket.AssignedToName || undefined,
    voteCount: apiTicket.VoteCount ?? 0,
    userHasVoted: apiTicket.UserHasVoted ?? false,
    resolutionNote: apiTicket.ResolutionNote || null,
  }
}

// The backend ships a small, configurable set of statuses (verified live:
// "Otevřeno", "Probíhá", "Vyřešeno"). Match them to the UI status enum by
// title so colours/labels are correct, with a position-based fallback for
// custom statuses and a status-less ticket treated as freshly created ("new").
const TITLE_MATCHERS: Array<[RegExp, TicketStatus]> = [
  [/otev/i, 'open'],
  [/(vyřeš|resolved|hotov|dokon)/i, 'resolved'],
  [/(prob|řeš|resol.*progres|in.?progress)/i, 'in_progress'],
  [/(uzav|zavř|closed)/i, 'closed'],
  [/(nov|new)/i, 'new'],
]

function statusFromTitle(title: string): TicketStatus | null {
  for (const [re, status] of TITLE_MATCHERS) {
    if (re.test(title)) return status
  }
  return null
}

// IsClosed je autoritativní zdroj pravdy o tom, co je "uzavřeno/vyřešeno" —
// title-regex/position se používá jen k odlišení open/in_progress mezi
// neuzavřenými stavy (žádný takový příznak pro ně neexistuje).
export function resolveStatus(
  statusId: ApiTicket['StatusID'],
  statuses: ApiTicketStatus[],
): TicketStatus {
  if (!statusId.Valid) return 'new'

  const match = statuses.find(s => s.ID === statusId.Int32)
  if (!match) return 'open'
  if (match.IsClosed) return 'resolved'

  const byTitle = statusFromTitle(match.Title)
  if (byTitle && byTitle !== 'resolved' && byTitle !== 'closed') return byTitle

  // Fallback: pořadí mezi neuzavřenými stavy — první → open, ostatní → in_progress.
  const open = statuses.filter(s => !s.IsClosed).sort((a, b) => a.Position - b.Position)
  const idx = open.findIndex(s => s.ID === match.ID)
  return idx <= 0 ? 'open' : 'in_progress'
}

// Resolves the server StatusID that corresponds to a UI status, used when
// changing a ticket's status. Returns undefined when no server status maps to
// the target (e.g. "new", which the server represents as a null StatusID).
export function statusIdForUiStatus(
  target: TicketStatus,
  statuses: ApiTicketStatus[],
): number | undefined {
  if (target === 'resolved' || target === 'closed') {
    const closed = statuses.filter(s => s.IsClosed).sort((a, b) => a.Position - b.Position)
    return closed[0]?.ID
  }

  const open = statuses.filter(s => !s.IsClosed).sort((a, b) => a.Position - b.Position)
  const direct = open.find(s => statusFromTitle(s.Title) === target)
  if (direct) return direct.ID

  if (target === 'open') return open[0]?.ID
  if (target === 'in_progress' && open.length > 1) return open[1]?.ID
  return undefined
}
