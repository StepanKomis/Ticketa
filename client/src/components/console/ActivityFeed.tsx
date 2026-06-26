import { ActivityEvent } from '../../types/ticket'
import './ActivityFeed.css'

const ACTION_LABELS: Record<ActivityEvent['action'], string> = {
  created:     'nový požadavek',
  in_progress: 'přesunuto na Řeší se',
  resolved:    'vyřešeno',
}

function relativeTime(date: Date): string {
  const diffMs = Date.now() - date.getTime()
  const mins  = Math.floor(diffMs / 60_000)
  const hours = Math.floor(diffMs / 3_600_000)
  const days  = Math.floor(diffMs / 86_400_000)
  if (mins < 60)  return `${mins} min`
  if (hours < 24) return `${hours} h`
  return `${days} d`
}

interface Props {
  events: ActivityEvent[]
}

export default function ActivityFeed({ events }: Props) {
  return (
    <section className="activityFeed" aria-label="Nedávná aktivita">
      <h2 className="activityFeed__heading">Nedávná aktivita</h2>
      {events.length === 0 ? (
        <p className="activityFeed__empty">Zatím žádná aktivita.</p>
      ) : (
        <ol className="activityFeed__list">
          {events.map(ev => (
            <li key={ev.id} className={`activityFeed__item activityFeed__item--${ev.action}`}>
              <span className="activityFeed__dot" aria-hidden="true" />
              <p className="activityFeed__text">
                <span className="activityFeed__title">{ev.title}</span>
                {' — '}{ACTION_LABELS[ev.action]}{' '}
                <span className="activityFeed__ticketId">{ev.ticketId}</span>
              </p>
              <span className="activityFeed__time">{relativeTime(ev.occurredAt)}</span>
            </li>
          ))}
        </ol>
      )}
    </section>
  )
}
