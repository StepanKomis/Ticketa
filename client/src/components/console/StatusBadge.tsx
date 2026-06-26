import { TicketStatus } from '../../types/ticket'
import './StatusBadge.css'

const LABELS: Record<TicketStatus, string> = {
  new:         'Nový',
  open:        'Otevřený',
  in_progress: 'Řeší se',
  resolved:    'Vyřešený',
  closed:      'Uzavřený',
}

// Small leading glyphs matching the design-system status swatches.
function StatusIcon({ status }: { status: TicketStatus }) {
  const common = {
    width: 10,
    height: 10,
    viewBox: '0 0 12 12',
    fill: 'none',
    xmlns: 'http://www.w3.org/2000/svg',
    'aria-hidden': true,
  } as const
  switch (status) {
    case 'in_progress':
      return (
        <svg {...common}>
          <circle cx="6" cy="6" r="4.5" stroke="currentColor" strokeWidth="1.2" />
          <path d="M6 3.5V6l1.8 1" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" />
        </svg>
      )
    case 'resolved':
    case 'closed':
      return (
        <svg {...common}>
          <circle cx="6" cy="6" r="4.5" stroke="currentColor" strokeWidth="1.2" />
          <path d="m4 6 1.4 1.5L8 4.5" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      )
    default:
      return (
        <svg {...common}>
          <circle cx="6" cy="6" r="4.5" stroke="currentColor" strokeWidth="1.2" />
        </svg>
      )
  }
}

interface Props {
  status: TicketStatus
}

export default function StatusBadge({ status }: Props) {
  return (
    <span className={`statusBadge statusBadge--${status}`}>
      <StatusIcon status={status} />
      {LABELS[status]}
    </span>
  )
}
