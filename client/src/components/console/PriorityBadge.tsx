import { TicketPriority } from '../../types/ticket'
import './PriorityBadge.css'

const LABELS: Record<TicketPriority, string> = {
  urgent: 'Urgentní',
  high:   'Vysoká',
  medium: 'Střední',
  low:    'Nízká',
}

interface Props {
  priority: TicketPriority
  pendingApproval?: boolean
}

export default function PriorityBadge({ priority, pendingApproval }: Props) {
  return (
    <span className={`priorityBadge priorityBadge--${priority}`}>
      {LABELS[priority]}
      {pendingApproval && <span className="priorityBadge__pending">· ke schválení</span>}
    </span>
  )
}
