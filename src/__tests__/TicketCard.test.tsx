import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import TicketCard from '../components/console/TicketCard'
import { Ticket } from '../types/ticket'

jest.mock('../hooks/useTickets', () => ({
  useVoteTicket:   () => ({ mutate: jest.fn(), isPending: false }),
  useUnvoteTicket: () => ({ mutate: jest.fn(), isPending: false }),
}))

const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
const renderCard = (ui: React.ReactElement) =>
  render(ui, {
    wrapper: ({ children }) => (
      <QueryClientProvider client={qc}>
        <MemoryRouter>{children}</MemoryRouter>
      </QueryClientProvider>
    ),
  })

const BASE_TICKET: Ticket = {
  id: 'TK-9999',
  title: 'Test tiket – projektor nefunguje',
  category: 'AV / Hardware',
  location: 'Místnost 101',
  status: 'open',
  priority: 'high',
  assigneeName: 'Adam Beneš',
  createdAt: new Date(Date.now() - 2 * 3_600_000),
  voteCount: 0,
  userHasVoted: false,
}

describe('TicketCard', () => {
  it('renders title, id, category and location', () => {
    renderCard(<TicketCard ticket={BASE_TICKET} />)
    expect(screen.getByText('Test tiket – projektor nefunguje')).toBeInTheDocument()
    expect(screen.getByText('TK-9999')).toBeInTheDocument()
    expect(screen.getByText('AV / Hardware')).toBeInTheDocument()
    expect(screen.getByText('Místnost 101')).toBeInTheDocument()
  })

  it('renders assignee initials', () => {
    renderCard(<TicketCard ticket={BASE_TICKET} />)
    expect(screen.getByText('AB')).toBeInTheDocument()
  })

  it('shows "Zahájit" action button for new status', () => {
    const ticket: Ticket = { ...BASE_TICKET, status: 'new' }
    const onAction = jest.fn()
    renderCard(<TicketCard ticket={ticket} onAction={onAction} />)
    const btn = screen.getByRole('button', { name: 'Zahájit' })
    fireEvent.click(btn)
    expect(onAction).toHaveBeenCalledWith(ticket)
  })

  it('shows "Vyřešit" action button for in_progress status', () => {
    const ticket: Ticket = { ...BASE_TICKET, status: 'in_progress' }
    const onAction = jest.fn()
    renderCard(<TicketCard ticket={ticket} onAction={onAction} />)
    expect(screen.getByRole('button', { name: 'Vyřešit' })).toBeInTheDocument()
  })

  it('shows no Zahájit/Vyřešit action button for resolved status', () => {
    const ticket: Ticket = { ...BASE_TICKET, status: 'resolved' }
    renderCard(<TicketCard ticket={ticket} onAction={() => {}} />)
    expect(screen.queryByRole('button', { name: 'Zahájit' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Vyřešit' })).not.toBeInTheDocument()
  })

  it('renders without assignee when assigneeName is undefined', () => {
    const ticket: Ticket = { ...BASE_TICKET, assigneeName: undefined }
    renderCard(<TicketCard ticket={ticket} />)
    expect(screen.queryByText('AB')).not.toBeInTheDocument()
  })
})
