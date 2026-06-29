export type TicketStatus = 'new' | 'open' | 'in_progress' | 'resolved' | 'closed'

export type TicketPriority = 'urgent' | 'high' | 'medium' | 'low'

export type TicketCategory =
  | 'AV / Hardware'
  | 'Síť / Internet'
  | 'Nábytek'
  | 'Budova / Prostory'
  | 'Účty / Přístupy'

export type UserRole = 'student' | 'staff' | 'maintainer' | 'admin' | 'pending'

export interface Assignee {
  initials: string
  name: string
}

export interface Ticket {
  id: string
  title: string
  status: TicketStatus
  createdAt: Date
  updatedAt?: Date
  resolvedAt?: Date
  body?: string
  authorId?: number
  authorName?: string
  priority?: TicketPriority
  requestedPriority?: TicketPriority | null
  category?: TicketCategory
  location?: string
  assigneeId?: number | null
  assigneeName?: string
  voteCount?: number
  userHasVoted?: boolean
  resolutionNote?: string | null
  isClosed?: boolean
}

// Activity is derived from real ticket data (no fabricated actors): a recent
// ticket and the state it is currently in.
export interface ActivityEvent {
  id: string
  ticketId: string
  title: string
  action: 'created' | 'in_progress' | 'resolved'
  occurredAt: Date
}
