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

export const PRIORITY_OPTIONS: { value: TicketPriority; label: string }[] = [
  { value: 'low',    label: 'Nízká' },
  { value: 'medium', label: 'Střední' },
  { value: 'high',   label: 'Vysoká' },
  { value: 'urgent', label: 'Urgentní' },
]
