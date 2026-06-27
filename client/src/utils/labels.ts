import type { TicketStatus, TicketPriority, UserRole } from '../types/ticket'

export const ROLE_LABELS: Record<UserRole, string> = {
  admin:      'Admin',
  staff:      'Učitel',
  maintainer: 'Údržbář',
  student:    'Student',
  pending:    'Čekající',
}

export const STATUS_LABELS: Record<TicketStatus, string> = {
  new:         'Nový',
  open:        'Otevřený',
  in_progress: 'Řeší se',
  resolved:    'Vyřešený',
  closed:      'Uzavřený',
}

// Record lookup for badge components; PRIORITY_OPTIONS (below) is the ordered
// array used by <select> elements — same data, different shape.
export const PRIORITY_LABELS: Record<TicketPriority, string> = {
  low:    'Nízká',
  medium: 'Střední',
  high:   'Vysoká',
  urgent: 'Urgentní',
}

export const PRIORITY_OPTIONS: { value: TicketPriority; label: string }[] = [
  { value: 'low',    label: PRIORITY_LABELS.low },
  { value: 'medium', label: PRIORITY_LABELS.medium },
  { value: 'high',   label: PRIORITY_LABELS.high },
  { value: 'urgent', label: PRIORITY_LABELS.urgent },
]
