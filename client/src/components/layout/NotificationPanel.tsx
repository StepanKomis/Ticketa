import { useNavigate } from 'react-router-dom'
import type { ApiNotification } from '../../types/api'
import { smartTime } from '../../utils/time'
import './NotificationPanel.css'

interface Props {
  items: ApiNotification[]
  onClose: () => void
}

export default function NotificationPanel({ items, onClose }: Props) {
  const navigate = useNavigate()

  const handleItemClick = (n: ApiNotification) => {
    onClose()
    if (n.ticket_id != null) {
      navigate(`/tickets/${n.ticket_id}`)
    }
  }

  return (
    <>
      <div className="notifBackdrop" onClick={onClose} aria-hidden="true" />
      <div className="notifPanel" role="dialog" aria-label="Oznámení">
        <div className="notifPanel__header">
          <span className="notifPanel__title">Oznámení</span>
          <button type="button" className="notifPanel__close" onClick={onClose} aria-label="Zavřít">×</button>
        </div>
        <div className="notifPanel__list">
          {items.length === 0 ? (
            <p className="notifPanel__empty">Žádná oznámení.</p>
          ) : (
            items.map(n => (
              <button
                key={n.id}
                type="button"
                className={`notifPanel__item${!n.is_viewed ? ' notifPanel__item--unread' : ''}`}
                onClick={() => handleItemClick(n)}
              >
                <span className="notifPanel__itemText">{n.text}</span>
                <span className="notifPanel__itemTime">{smartTime(new Date(n.created_at))}</span>
              </button>
            ))
          )}
        </div>
      </div>
    </>
  )
}
