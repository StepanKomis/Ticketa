import { render, screen } from '@testing-library/react'
import ActivityFeed from '../components/console/ActivityFeed'
import { ActivityEvent } from '../types/ticket'

const EVENTS: ActivityEvent[] = [
  {
    id: 'e1',
    ticketId: 'TK-2041',
    title: 'Projektor v učebně 204 nejde zapnout',
    action: 'in_progress',
    occurredAt: new Date(Date.now() - 25 * 60_000),
  },
  {
    id: 'e2',
    ticketId: 'TK-2027',
    title: 'Tiskárna v kantýně nereaguje',
    action: 'resolved',
    occurredAt: new Date(Date.now() - 3 * 86_400_000),
  },
]

describe('ActivityFeed', () => {
  it('renders the section heading', () => {
    render(<ActivityFeed events={EVENTS} />)
    expect(screen.getByText('Nedávná aktivita')).toBeInTheDocument()
  })

  it('renders each ticket title and id', () => {
    render(<ActivityFeed events={EVENTS} />)
    expect(screen.getByText('Projektor v učebně 204 nejde zapnout')).toBeInTheDocument()
    expect(screen.getByText('TK-2041')).toBeInTheDocument()
    expect(screen.getByText('Tiskárna v kantýně nereaguje')).toBeInTheDocument()
    expect(screen.getByText('TK-2027')).toBeInTheDocument()
  })

  it('renders the action label for an in-progress event', () => {
    render(<ActivityFeed events={[EVENTS[0]]} />)
    expect(screen.getByText(/přesunuto na Řeší se/)).toBeInTheDocument()
  })

  it('renders an empty state when there is no activity', () => {
    render(<ActivityFeed events={[]} />)
    expect(screen.getByText('Zatím žádná aktivita.')).toBeInTheDocument()
  })
})
