import { Circle, Clock, CheckCircle } from 'lucide-react'
import Card from '../ui/Card'
import './StatCards.scss'

interface Props {
  open: number
  inProgress: number
  resolved: number
}

export default function StatCards({ open, inProgress, resolved }: Props) {
  const cards = [
    { key: 'open', label: 'Otevřené', value: open, icon: <Circle size={12} strokeWidth={1.4} /> },
    { key: 'in_progress', label: 'Řeší se', value: inProgress, icon: <Clock size={12} strokeWidth={1.4} /> },
    { key: 'resolved', label: 'Vyřešené', value: resolved, icon: <CheckCircle size={12} strokeWidth={1.4} /> },
  ] as const

  return (
    <div className="statCards">
      {cards.map(card => (
        <Card key={card.key} className="statCard" data-stat={card.key}>
          <div className="statCard__top">
            <span className="statCard__label">{card.label}</span>
            <span className="statCard__icon" aria-hidden="true">{card.icon}</span>
          </div>
          <span className="statCard__value">{card.value}</span>
        </Card>
      ))}
    </div>
  )
}
