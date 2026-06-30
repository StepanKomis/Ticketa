import { ActivityEvent } from '../../types/ticket'
import { relativeTime } from '../../utils/time'
import Card from '../ui/Card'
import './ActivityFeed.scss'

const ACTION_LABELS: Record<ActivityEvent['action'], string> = {
  created:     'nový požadavek',
  in_progress: 'přesunuto na Řeší se',
  resolved:    'vyřešeno',
}

interface Props {
  events: ActivityEvent[]
}

export default function ActivityFeed({ events }: Props) {
  return (
    <Card className="activityFeed" aria-label="Nedávná aktivita">
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
    </Card>
  )
}
