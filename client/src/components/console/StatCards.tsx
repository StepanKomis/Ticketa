import './StatCards.css'

const CircleIcon = () => (
  <svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <circle cx="6" cy="6" r="4.5" stroke="currentColor" strokeWidth="1.3" />
  </svg>
)

const ClockIcon = () => (
  <svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <circle cx="6" cy="6" r="4.5" stroke="currentColor" strokeWidth="1.3" />
    <path d="M6 3.5V6l1.8 1" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" />
  </svg>
)

const CheckIcon = () => (
  <svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
    <circle cx="6" cy="6" r="4.5" stroke="currentColor" strokeWidth="1.3" />
    <path d="m4 6 1.4 1.5L8 4.5" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
)

interface Props {
  open: number
  inProgress: number
  resolved: number
}

export default function StatCards({ open, inProgress, resolved }: Props) {
  const cards = [
    { key: 'open', label: 'Otevřené', value: open, icon: <CircleIcon /> },
    { key: 'in_progress', label: 'Řeší se', value: inProgress, icon: <ClockIcon /> },
    { key: 'resolved', label: 'Vyřešené', value: resolved, icon: <CheckIcon /> },
  ] as const

  return (
    <div className="statCards">
      {cards.map(card => (
        <div key={card.key} className="statCard" data-stat={card.key}>
          <div className="statCard__top">
            <span className="statCard__label">{card.label}</span>
            <span className="statCard__icon" aria-hidden="true">{card.icon}</span>
          </div>
          <span className="statCard__value">{card.value}</span>
        </div>
      ))}
    </div>
  )
}
