import type { TicketPriority } from '../../types/ticket'
import { PRIORITY_LABELS } from '../../utils/labels'
import './PriorityBadge.css'

interface Props {
  priority: TicketPriority
  pendingApproval?: boolean
}

export default function PriorityBadge({ priority, pendingApproval }: Props) {
  return (
    <span className={`priorityBadge priorityBadge--${priority}`}>
      {PRIORITY_LABELS[priority]}
      {pendingApproval && <span className="priorityBadge__pending">· ke schválení</span>}
    </span>
  )
}
