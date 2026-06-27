import { Link } from 'react-router-dom'
import { Ticket } from '../../types/ticket'
import { formatTicketId } from '../../utils/mappers'
import { relativeTime } from '../../utils/time'
import { initialsFromName } from '../../utils/avatar'
import { useVoteTicket, useUnvoteTicket } from '../../hooks/useTickets'
import StatusBadge from './StatusBadge'
import PriorityBadge from './PriorityBadge'
import './TicketCard.css'

const PinIcon = () => (
  <svg width="10" height="10" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <path d="M6 1.5a3.2 3.2 0 0 1 3.2 3.2c0 2.3-3.2 5.8-3.2 5.8s-3.2-3.5-3.2-5.8A3.2 3.2 0 0 1 6 1.5Z" stroke="currentColor" strokeWidth="1.1"/>
    <circle cx="6" cy="4.7" r="1.1" stroke="currentColor" strokeWidth="1.1"/>
  </svg>
)

const UpvoteIcon = () => (
  <svg width="10" height="10" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <path d="M6 2 L10.5 9 L1.5 9 Z" stroke="currentColor" strokeWidth="1.2" strokeLinejoin="round" fill="currentColor" fillOpacity="0" />
  </svg>
)

interface Props {
  ticket: Ticket
  onAction?: (ticket: Ticket) => void
  canAct?: (ticket: Ticket) => boolean
}

export default function TicketCard({ ticket, onAction, canAct }: Props) {
  const ticketId = Number(ticket.id)
  const vote   = useVoteTicket(ticketId)
  const unvote = useUnvoteTicket(ticketId)

  const actionLabel =
    ticket.status === 'new' ? 'Zahájit' :
    ticket.status === 'in_progress' ? 'Vyřešit' : null

  function handleVote(e: React.MouseEvent) {
    e.preventDefault()
    e.stopPropagation()
    if (ticket.userHasVoted) {
      unvote.mutate()
    } else {
      vote.mutate()
    }
  }

  const assigneeInitials = initialsFromName(ticket.assigneeName)

  return (
    <article className="ticketCard" data-testid="ticket-card">
      <Link
        to={`/tickets/${ticket.id}`}
        className="ticketCard__link"
        aria-label={`Otevřít tiket ${ticket.title}`}
      />
      <span className={`ticketCard__dot ticketCard__dot--${ticket.status}`} aria-hidden="true" />

      <div className="ticketCard__body">
        <h3 className="ticketCard__title">{ticket.title}</h3>
        <p className="ticketCard__meta">
          <span className="ticketCard__id">{formatTicketId(ticket.id)}</span>
          {ticket.category && (
            <>
              <span className="ticketCard__sep" aria-hidden="true">·</span>
              <span>{ticket.category}</span>
            </>
          )}
          {ticket.location && (
            <>
              <span className="ticketCard__sep" aria-hidden="true">·</span>
              <span className="ticketCard__loc"><PinIcon />{ticket.location}</span>
            </>
          )}
        </p>
      </div>

      <div className="ticketCard__aside">
        {ticket.priority && (
          <span
            className={`ticketCard__prioDot ticketCard__prioDot--${ticket.priority}`}
            aria-hidden="true"
          />
        )}
        {ticket.priority && (
          <span className="ticketCard__prioLabel">
            <PriorityBadge priority={ticket.priority} pendingApproval={!!ticket.requestedPriority} />
          </span>
        )}
        <span className="ticketCard__time">{relativeTime(ticket.createdAt)}</span>
        {assigneeInitials && (
          <span className="ticketCard__avatar" title={ticket.assigneeName} aria-label={ticket.assigneeName}>
            {assigneeInitials}
          </span>
        )}
        <button
          type="button"
          className={`ticketCard__vote${ticket.userHasVoted ? ' ticketCard__vote--active' : ''}`}
          onClick={handleVote}
          aria-label={ticket.userHasVoted ? 'Odebrat hlas' : 'Hlasovat pro důležitost'}
          disabled={vote.isPending || unvote.isPending}
        >
          <UpvoteIcon />
          <span>{ticket.voteCount ?? 0}</span>
        </button>
        <StatusBadge status={ticket.status} />
        {actionLabel && onAction && (!canAct || canAct(ticket)) && (
          <button type="button" className="ticketCard__action" onClick={() => onAction(ticket)}>
            {actionLabel}
          </button>
        )}
      </div>
    </article>
  )
}
