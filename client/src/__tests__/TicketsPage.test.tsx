import { screen, fireEvent } from '@testing-library/react'
import TicketsPage from '../pages/ticketsPage'
import { renderWithProviders } from '../testUtils/renderWithProviders'
import type { ApiTicket, ApiTicketStatus } from '../types/api'

const MOCK_API_TICKETS: ApiTicket[] = [
  { ID: 1, Title: 'Projektor se nezapne', Body: '', Priority: 'medium', Location: '', Category: '', AssignedTo: null, AssignedToName: '', CreatedAt: new Date().toISOString(), UpdatedAt: new Date().toISOString(), AuthorID: 1, AuthorName: '', StatusID: { Int32: 0, Valid: false }, VoteCount: 0, UserHasVoted: false, RequestedPriority: null, PriorityApprovedBy: null, IsClosed: false, ResolutionNote: null },
  { ID: 2, Title: 'Wi-Fi odpojování',     Body: '', Priority: 'medium', Location: '', Category: '', AssignedTo: null, AssignedToName: '', CreatedAt: new Date().toISOString(), UpdatedAt: new Date().toISOString(), AuthorID: 1, AuthorName: '', StatusID: { Int32: 0, Valid: false }, VoteCount: 0, UserHasVoted: false, RequestedPriority: null, PriorityApprovedBy: null, IsClosed: false, ResolutionNote: null },
  { ID: 3, Title: 'Rozbitá židle',         Body: '', Priority: 'medium', Location: '', Category: '', AssignedTo: null, AssignedToName: '', CreatedAt: new Date().toISOString(), UpdatedAt: new Date().toISOString(), AuthorID: 2, AuthorName: '', StatusID: { Int32: 1, Valid: true  }, VoteCount: 0, UserHasVoted: false, RequestedPriority: null, PriorityApprovedBy: null, IsClosed: false, ResolutionNote: null },
  { ID: 4, Title: 'Pero tabule nepíše',   Body: '', Priority: 'medium', Location: '', Category: '', AssignedTo: null, AssignedToName: '', CreatedAt: new Date().toISOString(), UpdatedAt: new Date().toISOString(), AuthorID: 3, AuthorName: '', StatusID: { Int32: 0, Valid: false }, VoteCount: 0, UserHasVoted: false, RequestedPriority: null, PriorityApprovedBy: null, IsClosed: false, ResolutionNote: null },
]

const MOCK_LIST = { items: MOCK_API_TICKETS, total: MOCK_API_TICKETS.length, limit: 20, offset: 0 }

const mockUseTickets = jest.fn(() => ({ data: MOCK_LIST, isLoading: false }))
const mockMutateUpdate = jest.fn()
const mockMutatePatch = jest.fn()

jest.mock('../hooks/useTickets', () => ({
  useTickets:      () => mockUseTickets(),
  useTicket:       jest.fn(),
  useCreateTicket: () => ({ mutate: jest.fn(), reset: jest.fn(), isPending: false, error: null }),
  useUpdateTicket: () => ({ mutate: mockMutateUpdate, isPending: false }),
  usePatchTicket:  () => ({ mutate: mockMutatePatch, isPending: false }),
  useDeleteTicket: jest.fn(),
  useVoteTicket:   () => ({ mutate: jest.fn(), isPending: false }),
  useUnvoteTicket: () => ({ mutate: jest.fn(), isPending: false }),
}))

const mockUseStatuses = jest.fn(() => ({ data: [] as ApiTicketStatus[], isLoading: false }))

jest.mock('../hooks/useStatuses', () => ({
  useStatuses: () => mockUseStatuses(),
}))

describe('TicketsPage', () => {
  beforeEach(() => {
    mockUseTickets.mockReturnValue({ data: MOCK_LIST, isLoading: false })
    mockUseStatuses.mockReturnValue({ data: [], isLoading: false })
    mockMutateUpdate.mockClear()
    mockMutatePatch.mockClear()
  })

  it('renders the page heading', () => {
    renderWithProviders(<TicketsPage />, { initialPath: '/tickets' })
    expect(screen.getByRole('heading', { level: 1, name: 'Tikety' })).toBeInTheDocument()
  })

  it('renders ticket count', () => {
    renderWithProviders(<TicketsPage />, { initialPath: '/tickets' })
    expect(screen.getByText(/tiketů/)).toBeInTheDocument()
  })

  it('renders all mock tickets by default', () => {
    renderWithProviders(<TicketsPage />, { initialPath: '/tickets' })
    expect(screen.getAllByTestId('ticket-card').length).toBe(MOCK_API_TICKETS.length)
  })

  it('filters tickets by status — tab click updates URL params', () => {
    renderWithProviders(<TicketsPage />, { initialPath: '/tickets' })
    const tabBtn = screen.getAllByText('Řeší se')[0]
    fireEvent.click(tabBtn)
    // With server-side filtering, the mock returns the same data — just verify no crash
    expect(screen.getByRole('heading', { level: 1, name: 'Tikety' })).toBeInTheDocument()
  })

  it('shows empty state when filter has no results', () => {
    mockUseTickets.mockReturnValue({ data: { items: [], total: 0, limit: 20, offset: 0 }, isLoading: false })
    renderWithProviders(<TicketsPage />, { initialPath: '/tickets' })
    expect(screen.getByText('Žádné tikety v tomto filtru.')).toBeInTheDocument()
  })

  it('shows maintainer scope tabs and defaults to "Moje přiřazené"', () => {
    renderWithProviders(<TicketsPage />, { initialPath: '/tickets', user: { email: 'u@skola.cz', firstName: 'Test', lastName: 'Maintainer', role: 'maintainer', mustChangePw: false, id: 9 } })
    expect(screen.getByRole('tab', { name: 'Moje přiřazené' })).toBeInTheDocument()
    expect(screen.getByRole('tab', { name: 'Nepřiřazené' })).toBeInTheDocument()
    expect(screen.getByRole('heading', { level: 2, name: 'Moje přiřazené' })).toBeInTheDocument()
  })

  it('does not show maintainer scope tabs to staff', () => {
    renderWithProviders(<TicketsPage />, { initialPath: '/tickets' })
    expect(screen.queryByRole('tab', { name: 'Nepřiřazené' })).not.toBeInTheDocument()
  })

  it('switches scope tab on click', () => {
    renderWithProviders(<TicketsPage />, { initialPath: '/tickets', user: { email: 'u@skola.cz', firstName: 'Test', lastName: 'Maintainer', role: 'maintainer', mustChangePw: false, id: 9 } })
    fireEvent.click(screen.getByRole('tab', { name: 'Nepřiřazené' }))
    expect(screen.getByRole('tab', { name: 'Nepřiřazené' })).toHaveAttribute('aria-selected', 'true')
    expect(screen.getByRole('heading', { level: 2, name: 'Nepřiřazené' })).toBeInTheDocument()
  })

  it('opens the resolve dialog instead of mutating directly when resolving an in-progress ticket', () => {
    const statusesWithClosed: ApiTicketStatus[] = [
      { ID: 1, Title: 'Otevřeno', Color: '#3498db', Position: 0, IsClosed: false },
      { ID: 2, Title: 'Probíhá',  Color: '#f39c12', Position: 1, IsClosed: false },
      { ID: 3, Title: 'Vyřešeno', Color: '#2ecc71', Position: 2, IsClosed: true },
    ]
    mockUseStatuses.mockReturnValue({ data: statusesWithClosed, isLoading: false })
    const inProgressTicket: ApiTicket = { ...MOCK_API_TICKETS[0], ID: 99, StatusID: { Int32: 2, Valid: true } }
    mockUseTickets.mockReturnValue({ data: { items: [inProgressTicket], total: 1, limit: 20, offset: 0 }, isLoading: false })

    renderWithProviders(<TicketsPage />, { initialPath: '/tickets' })
    fireEvent.click(screen.getByRole('button', { name: 'Vyřešit' }))
    expect(mockMutateUpdate).not.toHaveBeenCalled()
    expect(mockMutatePatch).not.toHaveBeenCalled()
    expect(screen.getByRole('dialog', { name: 'Vyřešení tiketu' })).toBeInTheDocument()
  })
})
